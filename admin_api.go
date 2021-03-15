package prago

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago/utils"
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

	bindRelationAPI(app)
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
		user := request.GetUser()

		if !app.Authorize(user, resource.canEdit) {
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
			user := request.GetUser()

			file, err := app.UploadFile(v, &user, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func bindRelationAPI(app *App) {

	app.adminController.get(app.getAdminURL("_api/search/:resourceName"), func(request Request) {
		user := request.GetUser()
		resourceName := request.Params().Get("resourceName")
		q := request.Params().Get("q")

		usedIDs := map[int64]bool{}

		resource, found := app.resourceNameMap[resourceName]
		if !found {
			render404(request)
			return
		}

		if !app.Authorize(user, resource.canView) {
			render403(request)
			return
		}

		ret := []viewRelationData{}

		id, err := strconv.Atoi(q)
		if err == nil {
			var item interface{}
			resource.newItem(&item)
			err := app.Query().WhereIs("id", id).Get(item)
			if err == nil {
				relationItem := resource.itemToRelationData(item, user, nil)
				if relationItem != nil {
					//relationItem.Description = utils.Crop(relationItem.Description, 200)
					usedIDs[relationItem.ID] = true
					ret = append(ret, *relationItem)
				}
			}
		}

		filter := "%" + q + "%"
		for _, v := range []string{"name", "description"} {
			field := resource.fieldMap[v]
			if field == nil {
				continue
			}
			var items interface{}
			resource.newArrayOfItems(&items)
			err := app.Query().Limit(5).Where(v+" LIKE ?", filter).Get(items)
			if err == nil {
				itemsVal := reflect.ValueOf(items).Elem()
				for i := 0; i < itemsVal.Len(); i++ {
					var item interface{}
					item = itemsVal.Index(i).Interface()
					viewItem := resource.itemToRelationData(item, user, nil)
					if viewItem != nil && usedIDs[viewItem.ID] == false {
						usedIDs[viewItem.ID] = true
						ret = append(ret, *viewItem)
					}
				}
			}
		}

		if len(ret) > 5 {
			ret = ret[0:5]
		}

		for k := range ret {
			ret[k].Description = utils.Crop(ret[k].Description, 100)
		}

		request.RenderJSON(ret)
	})
}

func bindRelationListAPI(app *App) {
	app.API("relationlist").Method("POST").Handler(generateRelationListAPIHandler(app))
}
