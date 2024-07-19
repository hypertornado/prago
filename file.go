package prago

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

// File is structure representing files in admin
type File struct {
	ID          int64     `prago-order-desc:"true" prago-preview:"true"`
	UID         string    `prago-unique:"true" prago-preview:"true" prago-type:"cdnfile"`
	Name        string    `prago-can-edit:"nobody"`
	Description string    `prago-type:"text" prago-preview:"true"`
	User        int64     `prago-type:"relation" prago-can-edit:"nobody"`
	Width       int64     `prago-can-edit:"nobody" prago-preview:"true"`
	Height      int64     `prago-can-edit:"nobody" prago-preview:"true"`
	CreatedAt   time.Time `prago-preview:"true"`
	UpdatedAt   time.Time `prago-preview:"true"`
}

var filesCDN cdnclient.CDNAccount

func initCDN(app *App) {
	filesCDN = cdnclient.NewCDNAccount(
		app.mustGetSetting("cdn_url"),
		app.mustGetSetting("cdn_account"),
		app.mustGetSetting("cdn_password"),
	)
}

func (app *App) thumb(ids string) string {
	if ids == "" {
		return ""
	}
	for _, v := range strings.Split(ids, ",") {
		image := Query[File](app).Is("uid", v).First()
		if image != nil && image.IsImage() {
			return image.GetSmall()
		}
	}
	return ""
}

func (app *App) largeImage(ctx context.Context, ids string) string {
	if ids == "" {
		return ""
	}
	for _, v := range strings.Split(ids, ",") {
		image := Query[File](app).Context(ctx).Is("uid", v).First()
		if image != nil && image.IsImage() {
			return image.GetLarge()
		}
	}
	return ""
}

// GetFiles gets files from app
func (app *App) GetFiles(ctx context.Context, ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		if v == "" {
			continue
		}
		image := Query[File](app).Context(ctx).Is("uid", v).First()
		if image != nil {
			files = append(files, image)
		}
	}
	return files
}

func (app *App) UploadFile(fileHeader *multipart.FileHeader, request *Request, description string) (*File, error) {
	fileName := prettyFilename(fileHeader.Filename)
	file := File{}
	file.Name = fileName

	openedFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening multipart file: %s", err)
	}
	defer openedFile.Close()

	uploadData, err := filesCDN.UploadFile(openedFile, file.getExtension())
	if err != nil {
		return nil, fmt.Errorf("uploading multipart file: %s", err)
	}

	file.Width = uploadData.Width
	file.Height = uploadData.Height

	file.UID = uploadData.UUID

	file.User = request.UserID()
	file.Description = description
	err = CreateItemWithContext(request.Request().Context(), app, &file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}

// UpdateMetadata updates metadata of file
func (f *File) updateMetadata() error {
	metadata, err := filesCDN.GetMetadata(f.UID)
	if err != nil {
		return err
	}

	f.Width = metadata.Width
	f.Height = metadata.Height
	return nil
}

func (f File) getExtension() string {
	extension := filepath.Ext(f.Name)
	extension = strings.Replace(extension, ".", "", -1)
	return extension
}

func (app *App) initFilesResource() {
	initCDN(app)

	resource := NewResource[File](app)
	resource.Name(
		messages.GetNameFunction("admin_file"),
		messages.GetNameFunction("admin_files"),
	)
	app.FilesResource = resource

	resource.PermissionCreate(nobodyPermission)

	initFilesAPI(resource)

	resource.Field("uid").Name(messages.GetNameFunction("admin_file"))
	resource.Field("width").Name(messages.GetNameFunction("width"))
	resource.Field("height").Name(messages.GetNameFunction("height"))

	resource.Icon("glyphicons-basic-37-file.svg")

	app.addCommand("files", "metadata").
		Callback(func() {
			files := Query[File](app).List()
			for _, v := range files {
				err := v.updateMetadata()
				if err != nil {
					fmt.Println("error while updating metadata: ", v.ID, err)
					continue
				}
				f := *v
				if UpdateItem(app, &f) != nil {
					fmt.Println("error while saving file: ", v.ID)
				} else {
					fmt.Println("saved ok: ", v.ID, v.Width, v.Height)
				}
			}
		})

	app.ListenActivity(func(activity Activity) {
		if activity.ActivityType == "delete" && activity.ResourceID == resource.id {
			file := Query[File](app).ID(activity.ID)
			err := filesCDN.DeleteFile(file.UID)
			if err != nil {
				app.Log().Printf("deleting CDN: %s\n", err)
			}
		}
	})

	ResourceFormAction[File](app, "upload",
		func(f *Form, r *Request) {
			f.AddFileInput("file", messages.Get(r.Locale(), "admin_file"))
			f.AddTextareaInput("description", messages.Get(r.Locale(), "Description"))
			f.AddSubmit(messages.Get(r.Locale(), "admin_save"))
		},
		func(vc ValidationContext) {
			multipartFiles := vc.Request().Request().MultipartForm.File["file"]
			if len(multipartFiles) != 1 {
				vc.AddItemError("file", messages.Get(vc.Locale(), "admin_validation_not_empty"))
			}
			if vc.Valid() {
				fileData, err := app.UploadFile(multipartFiles[0], vc.Request(), vc.GetValue("description"))
				if err != nil {
					vc.AddError(err.Error())
				} else {
					vc.Validation().RedirectionLocation = fmt.Sprintf("/admin/file/%d", fileData.ID)
				}
			}
		},
	).setPriority(1000000).Permission(resource.canUpdate).Name(unlocalized("Nahrát soubor"))

	ActionResourcePlain[File](app, "getcdnurl", func(request *Request) {
		uuid := request.Param("uuid")
		size := request.Param("size")

		files := app.GetFiles(request.r.Context(), uuid)
		if len(files) == 0 {
			panic("can't find file")
		}
		file := files[0]

		redirectURL := filesCDN.GetImageURL(uuid, file.Name, size)
		request.Redirect(redirectURL)
	}).Permission(sysadminPermission).Method("POST")
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
	request.WriteJSON(200, responseData)
}

func (f *File) GetLarge() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "1000")
}

func (f *File) GetGiant() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "2500")
}

func (f *File) GetMedium() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "400")
}

func (f *File) GetSmall() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "200")
}

func (f *File) GetExactSize(width, height int) string {
	return filesCDN.GetImageURL(f.UID, f.Name, fmt.Sprintf("%dx%d", width, height))
}

func (f *File) GetOriginal() string {
	return filesCDN.GetFileURL(f.UID, f.Name)
}

func (f *File) GetMetadataPath() string {
	return filesCDN.MetadataPath(f.UID)
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}
