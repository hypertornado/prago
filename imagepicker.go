package prago

import (
	"encoding/json"
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
}

func (app *App) getImagePickerResponse(ids string) (ret *ImagePickerResponse) {

	ret = &ImagePickerResponse{}

	files := app.GetFiles(context.Background(), ids)
	for _, file := range files {
		image := &ImagePickerImage{
			UUID:             file.UID,
			ImageName:        file.Name,
			ImageDescription: file.Description,
			ViewURL:          fmt.Sprintf("/admin/file/%d", file.ID),
			EditURL:          fmt.Sprintf("/admin/file/%d/edit", file.ID),
			ThumbURL:         file.GetMedium(),
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
	return func(f *Field, userData UserData, value string) interface{} {
		//app := f.resource.app
		return imageFormData{
			MimeTypes: mimeTypes,
			//FileResponsesJSON: app.dataToFileResponseJSON(value),
		}
	}
}

func (app *App) dataToFileResponseJSON(data any) string {
	files := app.GetFiles(context.Background(), data.(string))
	fileResponse := getFileResponse(files)

	jsonResp, err := json.Marshal(fileResponse)
	must(err)
	return string(jsonResp)

}
