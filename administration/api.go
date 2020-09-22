package administration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/hypertornado/prago/utils"

	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
)

func bindAPI(a *Administration) {
	bindMarkdownAPI(a)
	bindRelationAPI(a)
}

func bindImageAPI(admin *Administration) {
	admin.AdminController.Get(admin.GetURL("file/uuid/:uuid"), func(request prago.Request) {
		var image File
		err := admin.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		request.Redirect(admin.GetURL(fmt.Sprintf("file/%d", image.ID)))
	})

	admin.AdminController.Post(admin.GetURL("_api/order/:resourceName"), func(request prago.Request) {
		resource := admin.getResourceByName(request.Params().Get("resourceName"))
		user := GetUser(request)

		if !admin.Authorize(user, resource.CanEdit) {
			panic("access denied")
		}

		if resource.OrderFieldName == "" {
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
			must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))
			must(resource.setOrderPosition(item, int64(i)))
			must(resource.Admin.Save(item))
		}
		request.RenderJSON(true)
	})

	admin.AdminController.Get(admin.GetURL("_api/image/thumb/:id"), func(request prago.Request) {
		var image File
		must(admin.Query().WhereIs("uid", request.Params().Get("id")).Get(&image))
		request.Redirect(image.GetMedium())
	})

	admin.AdminController.Get(admin.GetURL("_api/image/list"), func(request prago.Request) {
		basicUserAuthorize(request)
		var images []*File
		if len(request.Params().Get("ids")) > 0 {
			ids := strings.Split(request.Params().Get("ids"), ",")
			for _, v := range ids {
				var image File
				err := admin.Query().WhereIs("uid", v).Get(&image)
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
			q := admin.Query().WhereIs("filetype", "image").OrderDesc("createdat").Limit(10)
			if len(request.Params().Get("q")) > 0 {
				q = q.Where("name LIKE ? OR description LIKE ?", filter, filter)
			}
			must(q.Get(&images))
		}
		writeFileResponse(request, images)
	})

	admin.AdminController.Get(admin.GetURL("_api/imagedata/:uid"), func(request prago.Request) {
		basicUserAuthorize(request)
		var file File
		err := admin.Query().WhereIs("uid", request.Params().Get("uid")).Get(&file)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(file)
	})

	admin.AdminController.Post(admin.GetURL("_api/image/upload"), func(request prago.Request) {
		basicUserAuthorize(request)
		multipartFiles := request.Request().MultipartForm.File["file"]

		description := request.Params().Get("description")

		files := []*File{}

		for _, v := range multipartFiles {
			user := GetUser(request)

			file, err := admin.UploadFile(v, &user, description)
			if err != nil {
				panic(err)
			}
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func bindMarkdownAPI(admin *Administration) {
	admin.AdminController.Post(admin.GetURL("_api/markdown"), func(request prago.Request) {
		basicUserAuthorize(request)
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data))
	})
}

func bindRelationAPI(admin *Administration) {
	admin.AdminController.Get(admin.GetURL("_api/preview/:resourceName/:id"), func(request prago.Request) {
		resourceName := request.Params().Get("resourceName")
		idStr := request.Params().Get("id")

		user := GetUser(request)

		resource, found := admin.resourceNameMap[resourceName]
		if !found {
			render404(request)
			return
		}

		if !admin.Authorize(user, resource.CanView) {
			render403(request)
			return
		}

		var item interface{}
		resource.newItem(&item)
		err := admin.Query().WhereIs("id", idStr).Get(item)
		if err == ErrItemNotFound {
			render404(request)
			return
		}
		if err != nil {
			panic(err)
		}

		relationItem := resource.itemToRelationData(item, user, nil)
		request.RenderJSON(relationItem)
	})

	admin.AdminController.Get(admin.GetURL("_api/search/:resourceName"), func(request prago.Request) {
		user := GetUser(request)
		resourceName := request.Params().Get("resourceName")
		q := request.Params().Get("q")

		usedIDs := map[int64]bool{}

		resource, found := admin.resourceNameMap[resourceName]
		if !found {
			render404(request)
			return
		}

		if !admin.Authorize(user, resource.CanView) {
			render403(request)
			return
		}

		ret := []viewRelationData{}

		id, err := strconv.Atoi(q)
		if err == nil {
			var item interface{}
			resource.newItem(&item)
			err := admin.Query().WhereIs("id", id).Get(item)
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
			err := admin.Query().Limit(5).Where(v+" LIKE ?", filter).Get(items)
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
