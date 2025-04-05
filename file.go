package prago

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

// File is structure representing files in admin
type File struct {
	ID          int64  `prago-order-desc:"true"`
	UID         string `prago-unique:"true" prago-type:"cdnfile"`
	Name        string `prago-can-edit:"nobody"`
	Description string `prago-type:"text"`
	User        int64  `prago-type:"relation" prago-can-edit:"nobody"`
	Width       int64  `prago-can-edit:"nobody"`
	Height      int64  `prago-can-edit:"nobody"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

func (app *App) largeImage(ids string) string {
	if ids == "" {
		return ""
	}
	for _, v := range strings.Split(ids, ",") {
		image := Query[File](app).Is("uid", v).First()
		if image != nil && image.IsImage() {
			return image.GetLarge()
		}
	}
	return ""
}

func (app *App) GetFiles(ctx context.Context, ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		if v == "" {
			continue
		}
		file := Query[File](app).Context(ctx).Is("uid", v).First()
		if file != nil {
			files = append(files, file)
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

	ResourceAPI[File](app, "upload").Method("POST").Permission(resource.canUpdate).Handler(func(request *Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]
		description := request.Param("description")

		files := []*File{}
		for _, v := range multipartFiles {
			file, err := app.UploadFile(v, request, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}
		request.WriteJSON(200, getFileResponse(files))
	})

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

	ActionResourceForm[File](app, "upload",
		func(f *Form, r *Request) {
			f.AddFileInput("file", messages.Get(r.Locale(), "admin_file"))
			f.AddTextareaInput("description", messages.Get(r.Locale(), "Description"))
			f.AddSubmit(messages.Get(r.Locale(), "admin_save"))
		},
		func(vc FormValidation, request *Request) {
			multipartFiles := request.Request().MultipartForm.File["file"]
			if len(multipartFiles) != 1 {
				vc.AddItemError("file", messages.Get(request.Locale(), "admin_validation_not_empty"))
			}
			if vc.Valid() {
				fileData, err := app.UploadFile(multipartFiles[0], request, request.Param("description"))
				if err != nil {
					vc.AddError(err.Error())
				} else {
					vc.Redirect(fmt.Sprintf("/admin/file/%d", fileData.ID))
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

type fileResponse struct {
	FileURL      string
	UUID         string
	Name         string
	Description  string
	ThumbnailURL string
}

func getFileResponse(files []*File) []*fileResponse {
	responseData := []*fileResponse{}
	for _, v := range files {
		ir := &fileResponse{
			UUID:        v.UID,
			Name:        v.Name,
			Description: v.Description,
		}

		ir.FileURL = fmt.Sprintf("/admin/file/%d", v.ID)

		ir.ThumbnailURL = v.GetMedium()

		responseData = append(responseData, ir)
	}
	return responseData
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

func (f *File) getMetadataPath() string {
	return filesCDN.MetadataPath(f.UID)
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}

func (app *App) dataToFileResponseJSON(data any) string {
	files := app.GetFiles(context.Background(), data.(string))
	fileResponse := getFileResponse(files)

	jsonResp, err := json.Marshal(fileResponse)
	must(err)
	return string(jsonResp)

}

func fileViewDataSource(request *Request, field *Field, data interface{}) interface{} {
	return request.app.dataToFileResponseJSON(data)

}

type imageFormData struct {
	MimeTypes         string
	FileResponsesJSON string
}

func imageFormDataSource(mimeTypes string) func(*Field, UserData, string) interface{} {
	return func(f *Field, userData UserData, value string) interface{} {
		app := f.resource.app
		return imageFormData{
			MimeTypes:         mimeTypes,
			FileResponsesJSON: app.dataToFileResponseJSON(value),
		}
	}
}
