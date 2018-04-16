package administration

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
	"io"
	"mime/multipart"
	"os"
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

func initCDN(a *Administration) {
	cdnURL := a.App.Config.GetStringWithFallback("cdnURL", "https://www.prago-cdn.com")
	cdnAccount := a.App.Config.GetStringWithFallback("cdnAccount", a.HumanName)
	cdnPassword := a.App.Config.GetStringWithFallback("cdnPassword", "")
	filesCDN = cdnclient.NewCDNAccount(cdnURL, cdnAccount, cdnPassword)
}

func (a *Administration) GetFiles(ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		var image File
		err := a.Query().WhereIs("uid", v).Get(&image)
		if err == nil {
			files = append(files, &image)
		}
	}
	return files
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

func fileAfterFormCreated(f *Form, request prago.Request, newItem bool) *Form {
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

	openedFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening multipart file: %s", err)
	}
	defer openedFile.Close()

	uploadData, err := filesCDN.UploadFile(openedFile, file.GetExtension())
	if err != nil {
		return nil, fmt.Errorf("uploading multipart file: %s", err)
	}

	file.UID = uploadData.UUID
	return file, nil
}

func (f File) GetExtension() string {
	extension := filepath.Ext(f.Name)
	extension = strings.Replace(extension, ".", "", -1)
	return extension
}

func getOldRedirectParams(request prago.Request, admin *Administration) (uuid, name string, err error) {
	name = request.Params().Get("name")
	uuid = fmt.Sprintf("%s%s%s%s%s%s",
		request.Params().Get("a"),
		request.Params().Get("b"),
		request.Params().Get("c"),
		request.Params().Get("d"),
		request.Params().Get("e"),
		strings.Split(name, "-")[0],
	)

	var file File
	err = admin.Query().WhereIs("uid", uuid).Get(&file)
	if err != nil {
		return
	}
	name = file.Name
	return
}

func initFilesResource(resource *Resource) {
	a := resource.Admin
	initCDN(a)
	resource.AfterFormCreated = fileAfterFormCreated
	resource.Name = messages.Messages.GetNameFunction("admin_files")

	resource.ResourceController.AddBeforeAction(func(request prago.Request) {
		if request.Request().Method == "POST" && strings.HasSuffix(request.Request().URL.Path, "/delete") {
			idStr := request.Params().Get("id")
			id, err := strconv.Atoi(idStr)
			if err == nil {
				var file File
				must(a.Query().WhereIs("id", id).Get(&file))
				err = filesCDN.DeleteFile(file.UID)
				if err != nil {
					a.App.Log().Errorf("deleting CDN: %s", err)
				}
			}
		}
	})

	a.App.MainController().Get("/files/thumb/:size/:a/:b/:c/:d/:e/:name", func(request prago.Request) {
		uuid, name, err := getOldRedirectParams(request, a)
		if err != nil {
			panic(err)
		}

		var size int
		switch request.Params().Get("size") {
		case "large":
			size = 1000
		case "medium":
			size = 400
		case "small":
			size = 200
		default:
			panic("wrong size")
		}

		request.Redirect(filesCDN.GetImageURL(uuid, name, size))
	}, func(params map[string]string) bool {
		size := params["size"]
		if size == "large" || size == "medium" || size == "small" {
			return true
		}
		return false
	})

	a.App.MainController().Get("/files/original/:a/:b/:c/:d/:e/:name", func(request prago.Request) {
		uuid, name, err := getOldRedirectParams(request, a)
		must(err)
		request.Redirect(filesCDN.GetFileURL(uuid, name))
	})

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

	resource.ResourceController.Post(resource.GetURL(""), func(request prago.Request) {
		ValidateCSRF(request)

		multipartFiles := request.Request().MultipartForm.File["file"]
		if len(multipartFiles) != 1 {
			panic("must have 1 file selected")
		}

		file, err := uploadFile(multipartFiles[0], fileUploadPath)
		must(err)
		file.User = GetUser(request).ID
		file.Description = request.Params().Get("Description")
		must(a.Create(file))

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		request.Redirect(resource.GetURL(""))
	})

	resource.ResourceController.Get(resource.GetURL(":id/edit"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var file File
		must(a.Query().WhereIs("id", int64(id)).Get(&file))

		form := NewForm()
		form.Method = "POST"

		fi := form.AddTextInput("UUID", messages.Messages.Get(GetLocale(request), "UUID"))
		fi.Readonly = true
		fi.Value = file.UID

		fi = form.AddTextInput("Name", messages.Messages.Get(GetLocale(request), "Name"))
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
		var user User
		err = a.Query().WhereIs("id", file.User).Get(&user)
		if err == nil {
			fi.Value = fmt.Sprintf("%s (%d)", user.Name, user.ID)
		}

		fi = form.AddTextInput("uploadedAt", messages.Messages.Get(GetLocale(request), "Uploaded At"))
		fi.Readonly = true
		fi.Value = file.UpdatedAt.Format("2006-01-02 15:04:05")

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

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   a.getItemNavigation(*resource, user, &file, "edit"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

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
	request.RenderJSON(responseData)
}

//GetLarge file path
func (f *File) GetLarge() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 1000)
}

//GetMedium file path
func (f *File) GetMedium() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 400)
}

//GetSmall file path
func (f *File) GetSmall() string {
	return filesCDN.GetImageURL(f.UID, f.Name, 200)
}

//GetOriginal file path
func (f *File) GetOriginal() string {
	return filesCDN.GetFileURL(f.UID, f.Name)
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}