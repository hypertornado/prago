package prago

import (
	"fmt"

	"golang.org/x/net/context"
)

type ImagePickerResponse struct {
	Items []*ImagePickerImage
}

type ImagePickerImage struct {
	UUID             string
	ImageName        string
	ImageDescription string
	ViewURL          string
	EditURL          string
	ThumbURL         string

	Metadata [][2]string
}

func (app *App) getImagePickerResponse(ids string) (ret *ImagePickerResponse) {

	ret = &ImagePickerResponse{}

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

		image := &ImagePickerImage{
			UUID:             file.UID,
			ImageName:        file.Name,
			ImageDescription: file.Description,
			ViewURL:          fmt.Sprintf("/admin/file/%d", file.ID),
			EditURL:          fmt.Sprintf("/admin/file/%d/edit", file.ID),
			ThumbURL:         file.GetMedium(),
			Metadata:         metadata,
		}

		ret.Items = append(ret.Items, image)
	}

	return ret
}

func imagePickerAPIHandler(request *Request) any {
	return request.app.getImagePickerResponse(request.Param("ids"))
}

type imageFormData struct {
	MimeTypes         string
	FileResponsesJSON string
}

func imageFormDataSource(mimeTypes string) func(*Field, UserData, string) interface{} {
	return func(f *Field, userData UserData, value string) any {
		return imageFormData{
			MimeTypes: mimeTypes,
		}
	}
}
