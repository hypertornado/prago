package prago

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang-commonmark/markdown"
)

func (app *App) initAPI() {
	app.API("markdown").Method("POST").Handler(
		func(request Request) {
			basicUserAuthorize(request)
			data, err := ioutil.ReadAll(request.Request().Body)
			if err != nil {
				panic(err)
			}
			request.RenderJSON(markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data))
		},
	)

	bindRelationListAPI(app)
	bindImageAPI(app)
}

func bindImageAPI(app *App) {
	app.adminController.get(app.getAdminURL("file/uuid/:uuid"), func(request Request) {
		var image File
		err := app.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		request.Redirect(app.getAdminURL(fmt.Sprintf("file/%d", image.ID)))
	})

	app.adminController.post(app.getAdminURL("_api/order/:resourceName"), func(request Request) {
		resource := app.getResourceByName(request.Params().Get("resourceName"))
		user := request.getUser()

		if !app.authorize(user, resource.canEdit) {
			panic("access denied")
		}

		if resource.orderFieldName == "" {
			panic("can't order")
		}

		decoder := json.NewDecoder(request.Request().Body)
		var t = map[string][]int{}
		must(decoder.Decode(&t))

		order, ok := t["order"]
		if !ok {
			panic("wrong format")
		}

		for i, id := range order {
			var item interface{}
			resource.newItem(&item)
			must(resource.app.Query().WhereIs("id", int64(id)).Get(item))
			must(resource.setOrderPosition(item, int64(i)))
			must(resource.app.Save(item))
		}
		request.RenderJSON(true)
	})

	app.adminController.get(app.getAdminURL("_api/image/thumb/:id"), func(request Request) {
		var image File
		must(app.Query().WhereIs("uid", request.Params().Get("id")).Get(&image))
		request.Redirect(image.GetMedium())
	})

	app.adminController.get(app.getAdminURL("_api/image/list"), func(request Request) {
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
	})

	app.adminController.get(app.getAdminURL("_api/imagedata/:uid"), func(request Request) {
		basicUserAuthorize(request)
		var file File
		err := app.Query().WhereIs("uid", request.Params().Get("uid")).Get(&file)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(file)
	})

	app.adminController.post(app.getAdminURL("_api/image/upload"), func(request Request) {
		basicUserAuthorize(request)
		multipartFiles := request.Request().MultipartForm.File["file"]

		description := request.Params().Get("description")

		files := []*File{}

		for _, v := range multipartFiles {
			user := request.getUser()

			file, err := app.UploadFile(v, &user, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func bindRelationListAPI(app *App) {
	app.API("relationlist").Method("POST").Handler(generateRelationListAPIHandler(app))
}
