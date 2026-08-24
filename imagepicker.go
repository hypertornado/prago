package prago

import (
	"fmt"

	"golang.org/x/net/context"
)

type Image struct {
	UUID             string
	ImageName        string
	ImageDescription string
	ViewURL          string
	EditURL          string
	ThumbURL         string
	GiantURL         string

	Metadata [][2]string
}

func (app *App) GetImages(ids string) (ret []*Image) {

	files := app.GetFiles(context.Background(), ids)
	for _, file := range files {

		var metadata [][2]string
		metadata = append(metadata, [2]string{
			"ID",
			fmt.Sprintf("%d", file.ID),
		})
		metadata = append(metadata, [2]string{
			"UUID",
			file.UID,
		})

		if file.Width > 0 {
			metadata = append(metadata, [2]string{
				"Width",
				fmt.Sprintf("%d", file.Width),
			})
			metadata = append(metadata, [2]string{
				"Height",
				fmt.Sprintf("%d", file.Height),
			})
		}

		image := &Image{
			UUID:             file.UID,
			ImageName:        file.Name,
			ImageDescription: file.Description,
			ViewURL:          fmt.Sprintf("/admin/file/%d", file.ID),
			EditURL:          fmt.Sprintf("/admin/file/%d/edit", file.ID),
			ThumbURL:         file.GetMedium(),
			GiantURL:         file.GetGiant(),
			Metadata:         metadata,
		}

		ret = append(ret, image)
	}

	return ret
}

func imagePickerAPIHandler(request *Request) any {

	if !request.Authorize(request.app.FilesResource.canView) {
		panic("can't load image picker data")
	}

	return request.app.GetImages(request.Param("ids"))
}

type imageFormData struct {
	MimeTypes         string
	FileResponsesJSON string
}

func imageFormDataSource(mimeTypes string) func(*Field, UserData, string) any {
	return func(f *Field, userData UserData, value string) any {
		return imageFormData{
			MimeTypes: mimeTypes,
		}
	}
}

func imagePickerViewFieldContent(request *Request, field *Field, value any) *ViewContent {
	imageData := request.app.GetImages(value.(string))
	if len(imageData) == 0 {
		return nil
	}
	return &ViewContent{
		Images: imageData,
	}
}
