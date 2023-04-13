package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource[File]) {
	app := resource.data.app

	//TODO: remove this and use single details API
	resource.API("redirect-uuid/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := resource.Query(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	resource.API("redirect-thumb/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := resource.Query(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.Redirect(image.GetMedium())
	})

	resource.API("imagedata/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		file := resource.Query(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.WriteJSON(200, file)
	})

	resource.API("upload").Method("POST").Permission(resource.data.canUpdate).Handler(func(request *Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]
		description := request.Param("description")

		files := []*File{}

		for _, v := range multipartFiles {
			file, err := app.UploadFile(request.r.Context(), v, request, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}
		writeFileResponse(request, files)
	})
}
