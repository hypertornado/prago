package admin

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
	"github.com/renstrom/shortuuid"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	fileUploadPath   = ""
	fileDownloadPath = ""
)

var filesCDN cdnclient.CDNAccount

var thumbnailSizes = map[string][2]uint{
	"large":  {1000, 1000},
	"medium": {400, 400},
	"small":  {200, 200},
}

func initCDN(a *Admin) {
	cdnURL := a.App.Config.GetStringWithFallback("cdnURL", "https://prago-cdn.com")
	cdnAccount := a.App.Config.GetStringWithFallback("cdnAccount", a.AppName)
	cdnPassword := a.App.Config.GetStringWithFallback("cdnPassword", "")
	filesCDN = cdnclient.NewCDNAccount(cdnURL, cdnAccount, cdnPassword)
}

//File is structure representing files in admin
type File struct {
	ID          int64 `prago-order-desc:"true"`
	Name        string
	Description string `prago-type:"text" prago-preview:"true"`
	UID         string `prago-unique:"true" prago-preview:"true" prago-preview-type:"admin_image"`
	User        int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

//AdminName returns file admin name
func (File) AdminName(lang string) string { return messages.Messages.Get(lang, "admin_files") }

//AdminAfterFormCreated creates form for file upload
func (File) AdminAfterFormCreated(f *Form, request prago.Request, newItem bool) *Form {
	newForm := NewForm()
	newForm.Method = f.Method
	newForm.Action = f.Action
	if newItem {
		newForm.AddFileInput("file", messages.Messages.Get(GetLocale(request), "admin_file"))
		newForm.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
		newForm.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
	} else {
		newForm.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
	}
	AddCSRFToken(newForm, request)
	return newForm
}

func uploadFile(fileHeader *multipart.FileHeader, fileUploadPath string) (*File, error) {
	fileName := utils.PrettyFilename(fileHeader.Filename)
	file := &File{}
	file.Name = fileName
	file.UID = shortuuid.UUID()
	folderPath, filePath := file.getPath(fileUploadPath + "original")
	prago.Must(loadFile(folderPath, filePath, fileHeader))
	prago.Must(file.update(fileUploadPath))
	return file, nil
}

//InitResource of file
func (File) InitResource(a *Admin, resource *Resource) error {
	initCDN(a)

	filesExportCommand := a.App.CreateCommand("files:export", "export all files")
	a.App.AddCommand(filesExportCommand, func(app *prago.App) (err error) {
		fmt.Println("EXPORT COMMAND")

		var files []*File
		err = a.Query().Get(&files)
		if err != nil {
			return err
		}

		backupDir := fmt.Sprintf("%s/image-export-%s-%s", os.Getenv("HOME"), a.AppName, shortuuid.UUID())

		fmt.Println("Backing files to", backupDir)

		for _, v := range files {
			_, path := v.getPath(fileUploadPath + "original")
			fmt.Println(v.UID, path)

			dirPath := fmt.Sprintf("%s/%s/%s", backupDir, v.UID[0:2], v.UID[2:4])
			err = os.MkdirAll(dirPath, 0777)
			if err != nil {
				fmt.Println("mkdir error", err)
				continue
			}

			extension := filepath.Ext(v.Name)
			filePath := fmt.Sprintf("%s/%s%s", dirPath, v.UID, extension)

			cmd := exec.Command("cp", path, filePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("cp error", err)
				continue
			}
		}

		fmt.Println("Backed files to", backupDir)

		return nil
	})

	resource.DisplayInFooter = false
	resource.Pagination = 100

	fileUploadPath = a.App.Config.GetString("fileUploadPath")
	fileDownloadPath = a.App.Config.GetString("fileDownloadPath")

	if !strings.HasSuffix(fileUploadPath, "/") {
		fileUploadPath += "/"
	}
	if !strings.HasSuffix(fileDownloadPath, "/") {
		fileDownloadPath += "/"
	}

	bindImageAPI(a, fileDownloadPath)

	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		ValidateCSRF(request)

		multipartFiles := request.Request().MultipartForm.File["file"]
		if len(multipartFiles) != 1 {
			panic("must have 1 file selected")
		}

		file, err := uploadFile(multipartFiles[0], fileUploadPath)
		if err != nil {
			panic(err)
		}
		file.User = GetUser(request).ID
		file.Description = request.Params().Get("Description")
		prago.Must(a.Create(file))

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})

	resource.ResourceController.Get(a.GetURL(resource, ":id/edit"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var file File
		prago.Must(a.Query().WhereIs("id", int64(id)).Get(&file))

		form := NewForm()
		form.Method = "POST"

		fi := form.AddTextInput("Name", messages.Messages.Get(GetLocale(request), "Name"))
		fi.Readonly = true
		fi.Value = file.Name

		_, fileURL := file.getPath(fileDownloadPath + "original")

		fi = form.AddTextInput("url", messages.Messages.Get(GetLocale(request), "Url"))
		fi.Readonly = true
		fi.Value = fileURL
		fi.SubTemplate = "admin_item_link"

		fi = form.AddTextInput("uploadedBy", messages.Messages.Get(GetLocale(request), "Uploaded By"))
		fi.Readonly = true
		fi.Value = fmt.Sprintf("%d", file.User)

		if file.IsImage() {
			for _, v := range []string{"large", "medium", "small"} {
				fi = form.AddTextInput("thumb"+v, messages.Messages.Get(GetLocale(request), v))
				fi.Readonly = true
				_, path := file.getPath(fileDownloadPath + "thumb/" + v)
				fi.Value = path
				fi.SubTemplate = "admin_item_link"
			}
		}

		fi = form.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
		fi.Value = file.Description
		fi.Focused = true
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
		AddCSRFToken(form, request)

		request.SetData("admin_item", file)
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
	})

	return nil
}

func (f *File) getPath(prefix string) (folder, file string) {
	pathSeparator := "/"
	folder = prefix
	if len(folder) > 0 && !strings.HasSuffix(folder, pathSeparator) {
		folder += pathSeparator
	}

	if len(f.UID) < 7 {
		panic("too short uid")
	}

	uidPrefix := f.UID[0:5]
	folders := strings.Split(uidPrefix, "")
	folder += strings.Join(folders, pathSeparator)

	file = folder + pathSeparator + f.UID[5:] + "-" + f.Name
	return
}

func loadFile(folder, path string, header *multipart.FileHeader) error {
	if !strings.HasPrefix(path, folder) {
		return errors.New("folder path should be prefix of path")
	}
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		return fmt.Errorf("mkdirall : %s", err)
	}

	inFile, err := header.Open()
	if err != nil {
		return fmt.Errorf("opening header: %s", err)
	}

	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %s", err)
	}

	io.Copy(outFile, inFile)
	defer outFile.Close()
	defer inFile.Close()
	return nil
}

type imageResponse struct {
	ID          int64
	UID         string
	Name        string
	Description string
	Thumb       string
}

func writeFileResponse(request prago.Request, files []*File) {
	responseData := []*imageResponse{}
	for _, v := range files {
		ir := &imageResponse{
			ID:          v.ID,
			UID:         v.UID,
			Name:        v.Name,
			Description: v.Description,
		}

		_, fileURL := v.getPath(fileDownloadPath + "thumb/small")
		ir.Thumb = fileURL

		responseData = append(responseData, ir)
	}
	prago.WriteAPI(request, responseData, 200)
}

func bindImageAPI(a *Admin, fileDownloadPath string) {
	a.App.MainController().Get(a.Prefix+"/file/uuid/:uuid", func(request prago.Request) {
		var image File
		err := a.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		prago.Redirect(request,
			fmt.Sprintf("%s/file/%d/edit", a.Prefix, image.ID),
		)
	})

	a.App.MainController().Get(a.Prefix+"/_api/image/thumb/:id", func(request prago.Request) {
		var image File
		err := a.Query().WhereIs("uid", request.Params().Get("id")).Get(&image)
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, image.GetMedium())
	})

	a.App.MainController().Get(a.Prefix+"/_api/image/list", func(request prago.Request) {
		var images []*File

		if len(request.Params().Get("ids")) > 0 {
			ids := strings.Split(request.Params().Get("ids"), ",")
			for _, v := range ids {
				var image File
				err := a.Query().WhereIs("uid", v).Get(&image)
				if err == nil {
					images = append(images, &image)
				} else {
					if err != ErrItemNotFound {
						panic(err)
					}
				}
			}
		} else {
			filter := "%" + request.Params().Get("q") + "%"
			q := a.Query().WhereIs("filetype", "image").OrderDesc("createdat").Limit(10)
			if len(request.Params().Get("q")) > 0 {
				q = q.Where("name LIKE ? OR description LIKE ?", filter, filter)
			}
			prago.Must(q.Get(&images))
		}
		writeFileResponse(request, images)
	})

	a.AdminController.Post(a.Prefix+"/_api/image/upload", func(request prago.Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]

		description := request.Params().Get("description")

		files := []*File{}

		for _, v := range multipartFiles {
			file, err := uploadFile(v, fileUploadPath)
			if err != nil {
				panic(err)
			}
			file.User = GetUser(request).ID
			file.Description = description
			prago.Must(a.Create(file))
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func (f *File) update(fileUploadPath string) error {
	_, path := f.getPath(fileUploadPath + "original")

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	if f.IsImage() {
		for k, v := range thumbnailSizes {
			dirPath, filePath := f.getPath(fileUploadPath + "thumb/" + k)
			err := os.MkdirAll(dirPath, 0777)
			if err != nil {
				return fmt.Errorf("making dir: %s", err)
			}

			cmd := exec.Command("convert", path, "-auto-orient", "-thumbnail", fmt.Sprintf("%dx%d", v[0], v[1]), filePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("executing convert: %s", err)
			}
		}
	}
	return nil
}

//GetLarge file path
func (f *File) GetLarge() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 1000)
	//return f.getSize("large")
}

//GetMedium file path
func (f *File) GetMedium() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 400)
	//return f.getSize("medium")
}

//GetSmall file path
func (f *File) GetSmall() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 200)
	//return f.getSize("small")
}

//GetOriginal file path
func (f *File) GetOriginal() string {
	return filesCDN.GetFileURL(f.UID, f.Name)
	//_, path := f.getPath(fileDownloadPath + "original")
	//return path
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}

/*
func updateFiles(a *Admin) error {
	var files []*File
	a.Query().Get(&files)
	for _, file := range files {
		if file.IsImage() {
			err := file.update(fileUploadPath)
			if err != nil {
				return fmt.Errorf("updating file: %s", err)
			}
			err = a.Save(file)
			if err != nil {
				return fmt.Errorf("saving file: %s", err)
			}
		}
	}
	return nil
}
*/
