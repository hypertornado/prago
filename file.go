package prago

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

//File is structure representing files in admin
type File struct {
	ID          int64     `prago-order-desc:"true" prago-preview:"true"`
	UID         string    `prago-unique:"true" prago-preview:"true" prago-type:"cdnfile"`
	Name        string    `prago-can-edit:"nobody"`
	Description string    `prago-type:"text" prago-preview:"true"`
	User        int64     `prago-type:"relation" prago-preview:"true" prago-can-edit:"nobody"`
	Width       int64     `prago-can-edit:"nobody" prago-preview:"true"`
	Height      int64     `prago-can-edit:"nobody" prago-preview:"true"`
	CreatedAt   time.Time `prago-preview:"true"`
	UpdatedAt   time.Time `prago-preview:"true"`
}

var filesCDN cdnclient.CDNAccount

func initCDN(app *App) {
	cdnURL := app.ConfigurationGetStringWithFallback("cdnURL", "https://www.prago-cdn.com")
	cdnAccount := app.ConfigurationGetStringWithFallback("cdnAccount", app.codeName)
	cdnPassword := app.ConfigurationGetStringWithFallback("cdnPassword", "")
	filesCDN = cdnclient.NewCDNAccount(cdnURL, cdnAccount, cdnPassword)
}

func (app *App) thumb(ids string) string {
	if ids == "" {
		return ""
	}
	for _, v := range strings.Split(ids, ",") {
		image := app.FilesResource.Is("uid", v).First()
		if image != nil && image.isImage() {
			return image.GetSmall()
		}
	}
	return ""
}

//GetFiles gets files from app
func (app *App) GetFiles(ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		image := app.FilesResource.Is("uid", v).First()
		if image != nil {
			files = append(files, image)
		}
	}
	return files
}

func (app *App) UploadFile(fileHeader *multipart.FileHeader, user *user, description string) (*File, error) {
	fileName := prettyFilename(fileHeader.Filename)
	file := File{}
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

	file.Width = uploadData.Width
	file.Height = uploadData.Height

	file.UID = uploadData.UUID

	if user != nil {
		file.User = user.ID
	}
	file.Description = description
	err = app.create(&file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}

//UpdateMetadata updates metadata of file
func (f *File) updateMetadata() error {
	metadata, err := filesCDN.GetMetadata(f.UID)
	if err != nil {
		return err
	}

	f.Width = metadata.Width
	f.Height = metadata.Height
	return nil
}

//GetExtension gets file extension
func (f File) GetExtension() string {
	extension := filepath.Ext(f.Name)
	extension = strings.Replace(extension, ".", "", -1)
	return extension
}

func getOldRedirectParams(request *Request, app *App) (uuid, name string, err error) {
	name = request.Params().Get("name")
	uuid = fmt.Sprintf("%s%s%s%s%s%s",
		request.Params().Get("a"),
		request.Params().Get("b"),
		request.Params().Get("c"),
		request.Params().Get("d"),
		request.Params().Get("e"),
		strings.Split(name, "-")[0],
	)

	file := app.FilesResource.Is("uid", uuid).First()
	if file == nil {
		err = errors.New("no file with id found")
		return
	}
	name = file.Name
	return
}

func (app *App) initFilesResource() {
	initCDN(app)

	resource := NewResource[File](app)
	resource.Name(messages.GetNameFunction("admin_files"))
	app.FilesResource = resource

	resource.PermissionCreate(nobodyPermission)

	initFilesAPI(resource)

	resource.FieldName("uid", messages.GetNameFunction("admin_file"))
	resource.FieldName("width", messages.GetNameFunction("width"))
	resource.FieldName("height", messages.GetNameFunction("height"))

	app.addCommand("files", "metadata").
		Callback(func() {
			files := resource.Query().List()
			for _, v := range files {
				err := v.updateMetadata()
				if err != nil {
					fmt.Println("error while updating metadata: ", v.ID, err)
					continue
				}
				f := *v
				if resource.Update(&f) != nil {
					fmt.Println("error while saving file: ", v.ID)
				} else {
					fmt.Println("saved ok: ", v.ID, v.Width, v.Height)
				}
			}
		})

	app.ListenActivity(func(activity Activity) {
		if activity.ActivityType == "delete" && activity.ResourceID == resource.id {
			file := resource.Is("id", activity.ID).First()
			err := filesCDN.DeleteFile(file.UID)
			if err != nil {
				app.Log().Printf("deleting CDN: %s\n", err)
			}
		}
	})

	app.mainController.get("/files/thumb/:size/:a/:b/:c/:d/:e/:name", func(request *Request) {
		uuid, name, err := getOldRedirectParams(request, app)
		if err != nil {
			panic(err)
		}

		var size string
		switch request.Params().Get("size") {
		case "large":
			size = "1000"
		case "medium":
			size = "400"
		case "small":
			size = "200"
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

	app.mainController.get("/files/original/:a/:b/:c/:d/:e/:name", func(request *Request) {
		uuid, name, err := getOldRedirectParams(request, app)
		must(err)
		request.Redirect(filesCDN.GetFileURL(uuid, name))
	})

	newResourceFormAction(resource, "upload").priority().Permission(resource.canUpdate).Name(unlocalized("Nahrát soubor")).Form(func(f *Form, r *Request) {
		locale := r.user.Locale
		f.AddFileInput("file", messages.Get(locale, "admin_file"))
		f.AddTextareaInput("description", messages.Get(locale, "Description"))
		f.AddSubmit(messages.Get(locale, "admin_save"))
	}).Validation(func(vc ValidationContext) {
		multipartFiles := vc.Request().Request().MultipartForm.File["file"]
		if len(multipartFiles) != 1 {
			vc.AddItemError("file", messages.Get(vc.Locale(), "admin_validation_not_empty"))
		}
		if vc.Valid() {
			fileData, err := app.UploadFile(multipartFiles[0], vc.Request().user, vc.GetValue("description"))
			if err != nil {
				vc.AddError(err.Error())
			} else {
				vc.Validation().RedirectionLocaliton = fmt.Sprintf("/admin/file/%d", fileData.ID)
			}
		}
	})

	GetResource[File](app).Action("getcdnurl").Permission(sysadminPermission).Method("POST").Handler(
		func(request *Request) {
			uuid := request.Params().Get("uuid")
			size := request.Params().Get("size")

			files := app.GetFiles(uuid)
			if len(files) == 0 {
				panic("can't find file")
			}
			file := files[0]

			redirectURL := filesCDN.GetImageURL(uuid, file.Name, size)
			request.Redirect(redirectURL)
		},
	)
}

type imageResponse struct {
	ID          int64
	UID         string
	Name        string
	Description string
	Thumb       string
}

func writeFileResponse(request *Request, files []*File) {
	responseData := []*imageResponse{}
	for _, v := range files {
		ir := &imageResponse{
			ID:          v.ID,
			UID:         v.UID,
			Name:        v.Name,
			Description: v.Description,
		}

		ir.Thumb = v.GetMedium()

		responseData = append(responseData, ir)
	}
	request.RenderJSON(responseData)
}

//GetLarge file path
func (f *File) GetLarge() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "1000")
}

//GetGiant file path
func (f *File) GetGiant() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "2500")
}

//GetMedium file path
func (f *File) GetMedium() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "400")
}

//GetSmall file path
func (f *File) GetSmall() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "200")
}

//GetOriginal file path
func (f *File) GetOriginal() string {
	return filesCDN.GetFileURL(f.UID, f.Name)
}

//GetMetadataPath gets metadada file path
func (f *File) GetMetadataPath() string {
	return filesCDN.MetadataPath(f.UID)
}

//IsImage detects if file is image
func (f *File) IsImage() bool {
	return f.isImage()
}

func (f *File) isImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}
