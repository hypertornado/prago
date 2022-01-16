package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource[File]) {
	app := resource.app

	//TODO: remove this and use single details API
	resource.API("redirect-uuid/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := resource.Is("uid", request.Params().Get("uuid")).First()
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	resource.API("redirect-thumb/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := resource.Is("uid", request.Params().Get("uuid")).First()
		request.Redirect(image.GetMedium())
	})

	resource.API("imagedata/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		file := resource.Is("uid", request.Params().Get("uuid")).First()
		request.RenderJSON(file)
	})

	resource.API("upload").Method("POST").Permission(resource.canUpdate).Handler(func(request *Request) {
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
