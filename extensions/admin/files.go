package admin

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/utils"
	"github.com/renstrom/shortuuid"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	fileUploadPath   = ""
	fileDownloadPath = ""
)

var ThumbnailSizes = map[string][2]uint{
	"large":  {1000, 1000},
	"medium": {400, 400},
	"small":  {200, 200},
}

type File struct {
	ID          int64 `prago-order-desc:"true"`
	Name        string
	Description string `prago-type:"text"`
	UID         string `prago-unique:"true"`
	FileType    string
	Size        int64
	Width       int64
	Height      int64
	UserId      int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (File) AdminName(lang string) string { return messages.Messages.Get(lang, "admin_files") }

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

func UploadFile(fileHeader *multipart.FileHeader, fileUploadPath string) (*File, error) {
	fileName := utils.PrettyFilename(fileHeader.Filename)
	file := &File{}
	file.Name = fileName
	file.UID = shortuuid.UUID()
	folderPath, filePath := file.GetPath(fileUploadPath + "original")
	prago.Must(loadFile(folderPath, filePath, fileHeader))
	prago.Must(file.Update(fileUploadPath))
	return file, nil
}

func (File) AdminInitResource(a *Admin, resource *AdminResource) error {
	resource.Pagination = 100

	fileUploadPath = a.App.Config().GetString("fileUploadPath")
	fileDownloadPath = a.App.Config().GetString("fileDownloadPath")

	if !strings.HasSuffix(fileUploadPath, "/") {
		fileUploadPath += "/"
	}
	if !strings.HasSuffix(fileDownloadPath, "/") {
		fileDownloadPath += "/"
	}

	BindImageAPI(a, fileDownloadPath)

	resource.Actions["create"] = func(a *Admin, resource *AdminResource) {
		resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
			ValidateCSRF(request)

			multipartFiles := request.Request().MultipartForm.File["file"]
			if len(multipartFiles) != 1 {
				panic("must have 1 file selected")
			}

			file, err := UploadFile(multipartFiles[0], fileUploadPath)
			if err != nil {
				panic(err)
			}
			file.UserId = GetUser(request).ID
			file.Description = request.Params().Get("Description")
			prago.Must(resource.Create(file))

			FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
			prago.Redirect(request, a.Prefix+"/"+resource.ID)
		})
	}

	resource.Actions["detail"] = func(a *Admin, resource *AdminResource) {
		resource.ResourceController.Get(a.GetURL(resource, ":id"), func(request prago.Request) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			prago.Must(err)

			item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
			prago.Must(err)

			file := item.(*File)

			form := NewForm()
			form.Method = "POST"

			fi := form.AddTextInput("Name", messages.Messages.Get(GetLocale(request), "Name"))
			fi.Readonly = true
			fi.Value = file.Name

			_, fileUrl := file.GetPath(fileDownloadPath + "original")

			fi = form.AddTextInput("url", messages.Messages.Get(GetLocale(request), "Url"))
			fi.Readonly = true
			fi.Value = fileUrl

			fi = form.AddTextInput("size", messages.Messages.Get(GetLocale(request), "Size"))
			fi.Readonly = true
			fi.Value = fmt.Sprintf("%d", file.Size)

			fi = form.AddTextInput("uploadedBy", messages.Messages.Get(GetLocale(request), "Uploaded By"))
			fi.Readonly = true
			fi.Value = fmt.Sprintf("%d", file.UserId)

			fi = form.AddTextInput("fileType", messages.Messages.Get(GetLocale(request), "Type"))
			fi.Readonly = true
			fi.Value = file.FileType

			if file.IsImage() {
				fi = form.AddTextInput("width", messages.Messages.Get(GetLocale(request), "Width"))
				fi.Readonly = true
				fi.Value = fmt.Sprintf("%d", file.Width)

				fi = form.AddTextInput("height", messages.Messages.Get(GetLocale(request), "Height"))
				fi.Readonly = true
				fi.Value = fmt.Sprintf("%d", file.Height)

				for _, v := range []string{"large", "medium", "small"} {
					fi = form.AddTextInput("thumb"+v, messages.Messages.Get(GetLocale(request), v))
					fi.Readonly = true
					_, path := file.GetPath(fileDownloadPath + "thumb/" + v)
					fi.Value = path
				}

			}

			fi = form.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
			fi.Value = file.Description
			fi.Focused = true
			form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
			AddCSRFToken(form, request)

			request.SetData("admin_item", item)
			request.SetData("admin_form", form)
			request.SetData("admin_yield", "admin_edit")
			prago.Render(request, 200, "admin_layout")
		})
	}

	return nil
}

func (f *File) GetPath(prefix string) (folder, file string) {
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
		return err
	}

	inFile, err := header.Open()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}

	io.Copy(outFile, inFile)

	defer outFile.Close()
	defer inFile.Close()
	return nil
}

type ImageResponse struct {
	ID          int64
	UID         string
	Name        string
	Description string
	Thumb       string
}

func writeFileResponse(request prago.Request, files []*File) {
	responseData := []*ImageResponse{}
	for _, v := range files {
		ir := &ImageResponse{
			ID:          v.ID,
			UID:         v.UID,
			Name:        v.Name,
			Description: v.Description,
		}

		_, fileUrl := v.GetPath(fileDownloadPath + "thumb/small")
		ir.Thumb = fileUrl

		responseData = append(responseData, ir)
	}

	WriteApi(request, responseData, 200)

}

func BindImageAPI(a *Admin, fileDownloadPath string) {
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
					if err != ErrorNotFound {
						panic(err)
					}
				}
			}
		} else {
			filter := "%" + request.Params().Get("q") + "%"
			q := a.Query().WhereIs("filetype", "image").OrderDesc("createdat").Limit(10)
			if len(request.Params().Get("q")) > 0 {
				q = q.Where("name LIKE ?", filter)
			}
			prago.Must(q.Get(&images))
		}
		writeFileResponse(request, images)

	})

	a.AdminController.Post(a.Prefix+"/_api/image/upload", func(request prago.Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]

		files := []*File{}

		for _, v := range multipartFiles {
			file, err := UploadFile(v, fileUploadPath)
			if err != nil {
				panic(err)
			}
			file.UserId = GetUser(request).ID
			prago.Must(a.Create(file))
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func (f *File) Update(fileUploadPath string) error {
	_, path := f.GetPath(fileUploadPath + "original")

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	f.Size = stat.Size()

	if f.IsImage() {
		img, err := jpeg.Decode(file)
		if err != nil {
			return err
		}
		bounds := img.Bounds()
		f.Width = int64(bounds.Size().X)
		f.Height = int64(bounds.Size().Y)
		f.FileType = "image"

		for k, v := range ThumbnailSizes {
			dirPath, filePath := f.GetPath(fileUploadPath + "thumb/" + k)
			err := os.MkdirAll(dirPath, 0777)
			if err != nil {
				return err
			}

			cmd := exec.Command("convert", path, "-thumbnail", fmt.Sprintf("%dx%d", v[0], v[1]), filePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *File) getSize(size string) string {
	_, path := f.GetPath(fileDownloadPath + "thumb/" + size)
	return path
}

func (f *File) GetLarge() string {
	return f.getSize("large")
}

func (f *File) GetMedium() string {
	return f.getSize("medium")
}

func (f *File) GetSmall() string {
	return f.getSize("small")
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") {
		return true
	}
	return false
}

func UpdateFiles(a *Admin) error {
	var files []*File
	a.Query().Get(&files)
	for _, file := range files {
		fmt.Println(file.UID, file.Name)
		if file.IsImage() {
			err := file.Update(fileUploadPath)
			if err != nil {
				return err
			}
			err = a.Save(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
