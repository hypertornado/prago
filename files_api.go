package prago

import (
	"fmt"
)

func initFilesAPI(resource *Resource) {
	app := resource.app

	//TODO: remove this and use single details API
	resource.API("redirect-uuid/:uuid").Handler(func(request *Request) {
		var image File
		err := app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	resource.API("redirect-thumb/:uuid").Handler(func(request *Request) {
		var image File
		must(app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image))
		request.Redirect(image.GetMedium())
	})

	/*
		app.adminController.get(app.getAdminURL("_api/image/list"), func(request *Request) {
			basicUserAuthorize(request)
			var images []*File
			if len(request.Params().Get("ids")) > 0 {
				ids := strings.Split(request.Params().Get("ids"), ",")
				for _, v := range ids {
					var image File
					err := app.Query().WhereIs("uid", v).Get(&image)
					if err == nil {
						images = append(images, &image)
					} else {
						if err != ErrItemNotFound {
							panic(err)
						}
					}
				}
			} else {
				filter := "%" + request.Params().Get("q") + "%"
				q := app.Query().WhereIs("filetype", "image").OrderDesc("createdat").Limit(10)
				if len(request.Params().Get("q")) > 0 {
					q = q.Where("name LIKE ? OR description LIKE ?", filter, filter)
				}
				must(q.Get(&images))
			}
			writeFileResponse(request, images)
		})*/

	resource.API("imagedata/:uuid").Handler(func(request *Request) {
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
