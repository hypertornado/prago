package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource) {
	app := resource.app

	//TODO: remove this and use single details API
	resource.API("redirect-uuid/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var image File
		must(app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image))
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	resource.API("redirect-thumb/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var image File
		must(app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image))
		request.Redirect(image.GetMedium())
	})

	resource.API("imagedata/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		var file File
		must(app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&file))
		request.RenderJSON(file)
	})

	resource.API("upload").Method("POST").Permission(resource.canCreate).Handler(func(request *Request) {
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
