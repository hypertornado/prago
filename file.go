package prago

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"strings"
	"time"
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

}

func (app *App) afterInitFilesResource() {
	resource := app.FilesResource

	ResourceAPI[File](app, "upload").Method("POST").Permission(resource.canUpdate).Handler(func(request *Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]
		description := request.Param("description")

		var uuids []string

		for _, v := range multipartFiles {
			file, err := app.UploadFile(v, request, description)
			must(err)
			uuids = append(uuids, file.UID)
		}
		request.WriteJSON(200, uuids)
	})

	app.API("imagepicker").Permission(loggedPermission).HandlerJSON(imagePickerAPIHandler)

	resource.Field("uid").Name(messages.GetNameFunction("admin_file"))
	resource.Field("width").Name(messages.GetNameFunction("width"))
	resource.Field("height").Name(messages.GetNameFunction("height"))

	resource.Icon("glyphicons-basic-37-file.svg")

	ActionResourceItemForm(app, "download", func(file *File, form *Form, request *Request) {
		dataPaths := file.getCDNNamedDownloadPaths()

		var values [][2]string
		for _, v := range dataPaths {
			values = append(values, [2]string{
				v[0],
				v[0],
			})
		}

		form.AddSelect("typ", "Type", values).Value = "original"
		if file.IsImage() {
			form.AddTextInput("custom", "Custom size")
		}
		form.AddSubmit("Download")
	}, func(file *File, fv FormValidation, request *Request) {

		customSize := request.Param("custom")
		if customSize != "" {
			redirectURL := filesCDN.GetImageURL(file.UID, file.Name, customSize)
			fv.Redirect(redirectURL)
			return
		}

		dataPaths := file.getCDNNamedDownloadPaths()
		for _, v := range dataPaths {
			if v[0] == request.Param("typ") {
				fv.Redirect(v[1])
				return
			}
		}
		fv.AddError("No size selected")

	}).Name(unlocalized("Download")).Icon("glyphicons-basic-199-save.svg")

	ActionResourceItemUI(app, "metadata", func(file *File, request *Request) template.HTML {
		metadata, err := filesCDN.GetMetadata(file.UID)
		if err != nil {
			panic(err)
		}

		table := app.Table()

		table.Row(
			Cell("UUID"),
			Cell(metadata.UUID),
		)
		table.Row(
			Cell("Extension"),
			Cell(metadata.Extension),
		)
		table.Row(
			Cell("IsImage"),
			Cell(metadata.IsImage),
		)
		table.Row(
			Cell("Filesize"),
			Cell(metadata.Filesize),
		)
		table.Row(
			Cell("Width"),
			Cell(metadata.Width),
		)
		table.Row(
			Cell("Height"),
			Cell(metadata.Height),
		)

		return table.ExecuteHTML()

	}).Name(unlocalized("CDN Metadata")).Icon("glyphicons-basic-501-server.svg")

	ActionResourceItemUI(app, "connections", func(file *File, request *Request) template.HTML {
		table := app.Table()
		table.Header("Resource", "Field", "Item")

		var totalConnections = 0

		for _, resource := range app.resources {
			for _, field := range resource.fields {
				if field.fieldType.viewTemplate == "view_image" {
					query := fmt.Sprintf("SELECT id FROM %s WHERE %s LIKE ?", resource.id, field.id)

					rows, err := app.db.Query(query, "%"+file.UID+"%")
					if err != nil {
						log.Fatal(err)
					}
					defer rows.Close()

					for rows.Next() {
						var id int
						if err := rows.Scan(&id); err != nil {
							log.Fatal(err)
						}

						totalConnections++

						item := resource.query(context.Background()).ID(id)
						previewer := resource.previewer(request, item)

						if request.Authorize(resource.canView) && request.Authorize(field.canView) {
							table.Row(
								Cell(resource.singularName(request.Locale())).URL(fmt.Sprintf("/admin/%s", resource.id)),
								Cell(field.name(request.Locale())),
								Cell(previewer.Name()).URL(previewer.URL("")),
							)
						} else {
							table.Row(Cell("Not authorized"))
						}
					}

					if err := rows.Err(); err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		if totalConnections == 1 {
			table.AddFooterText(fmt.Sprintf("%d connection", totalConnections))
		} else {
			table.AddFooterText(fmt.Sprintf("%d connections", totalConnections))
		}

		return table.ExecuteHTML()
	}).Name(unlocalized("Connections")).Icon("glyphicons-basic-63-paperclip.svg")

	ItemStatistic(app, unlocalized("UUID"), app.FilesResource.canView, func(file *File) string {
		return file.UID
	})

	ItemStatistic(app, unlocalized("Connections"), app.FilesResource.canView, func(file *File) string {
		var totalConnections int64
		for _, resource := range app.resources {
			for _, field := range resource.fields {
				if field.fieldType.viewTemplate == "view_image" {
					query := fmt.Sprintf("SELECT id FROM %s WHERE %s LIKE ?", resource.id, field.id)

					rows, err := app.db.Query(query, "%"+file.UID+"%")
					if err != nil {
						log.Fatal(err)
					}
					defer rows.Close()

					for rows.Next() {
						var id int
						if err := rows.Scan(&id); err != nil {
							log.Fatal(err)
						}

						totalConnections++
					}

					if err := rows.Err(); err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		if totalConnections == 1 {
			return fmt.Sprintf("%d connection", totalConnections)
		} else {
			return fmt.Sprintf("%d connections", totalConnections)
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
}

type fileUploadResponse struct {
	FileURL      string
	UUID         string
	Name         string
	Description  string
	ThumbnailURL string
}

func getFileUploadResponses(files []*File) []*fileUploadResponse {
	responseData := []*fileUploadResponse{}
	for _, v := range files {
		ir := &fileUploadResponse{
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

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}

func fileViewDataSource(request *Request, field *Field, data any) any {
	outData := request.app.getImagePickerResponse(data.(string))
	jsonData, err := json.Marshal(outData)
	must(err)
	return string(jsonData)
}
