package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource) {
	app := resource.app

	//TODO: remove this and use single details API
	ResourceAPI[File](app, "redirect-uuid/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := Query[File](app).Context(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	ResourceAPI[File](app, "redirect-thumb/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		image := Query[File](app).Context(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.Redirect(image.GetMedium())
	})

	ResourceAPI[File](app, "imagedata/:uuid").Permission(loggedPermission).Handler(func(request *Request) {
		file := Query[File](app).Context(request.r.Context()).Is("uid", request.Param("uuid")).First()
		request.WriteJSON(200, file)
	})

	ResourceAPI[File](app, "upload").Method("POST").Permission(resource.canUpdate).Handler(func(request *Request) {
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
