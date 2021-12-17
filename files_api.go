package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource) {
	app := resource.app

	//TODO: remove this and use single details API
	resource.API("redirect-uuid/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var image File
		app.Is("uid", request.Params().Get("uuid")).MustGet(&image)
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	resource.API("redirect-thumb/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var image File
		app.Is("uid", request.Params().Get("uuid")).MustGet(&image)
		request.Redirect(image.GetMedium())
	})

	resource.API("imagedata/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var file File
		app.Is("uid", request.Params().Get("uuid")).MustGet(&file)
		request.RenderJSON(file)
	})

	resource.API("upload").Method("POST").Permission(resource.canEdit).Handler(func(request *Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]
		description := request.Params().Get("description")

		files := []*File{}

		for _, v := range multipartFiles {
			file, err := app.UploadFile(v, request.user, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}
		writeFileResponse(request, files)
	})
}
