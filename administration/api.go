package administration

import (
	"encoding/json"
	"fmt"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"reflect"
	"strings"
)

func bindAPI(a *Administration) {
	bindMarkdownAPI(a)
	bindListAPI(a)
	bindListResourceAPI(a)
	bindListResourceItemAPI(a)
}

func bindImageAPI(admin *Administration, fileDownloadPath string) {
	admin.AdminController.Get(admin.GetURL("file/uuid/:uuid"), func(request prago.Request) {
		var image File
		err := admin.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		request.Redirect(admin.GetURL(fmt.Sprintf("file/%d", image.ID)))
	})

	admin.AdminController.Get(admin.GetURL("_api/image/thumb/:id"), func(request prago.Request) {
		var image File
		must(admin.Query().WhereIs("uid", request.Params().Get("id")).Get(&image))
		request.Redirect(image.GetMedium())
	})

	admin.AdminController.Get(admin.GetURL("_api/image/list"), func(request prago.Request) {
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

	admin.AdminController.Post(admin.GetURL("_api/image/upload"), func(request prago.Request) {
		multipartFiles := request.Request().MultipartForm.File["file"]

		description := request.Params().Get("description")

		files := []*File{}

		for _, v := range multipartFiles {
			file, err := uploadFile(v, fileUploadPath)
			if err != nil {
				panic(err)
			}
			file.User = GetUser(request).ID
			file.Description = description
			must(admin.Create(file))
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func bindMarkdownAPI(admin *Administration) {
	admin.AdminController.Post(admin.GetURL("_api/markdown"), func(request prago.Request) {
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data))
	})
}

type resourceItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func bindListAPI(admin *Administration) {
	admin.AdminController.Post(admin.GetURL("_api/list/:name"), func(request prago.Request) {
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := admin.resourceNameMap[name]
		if !found {
			render404(request)
			return
		}

		if !admin.Authorize(*user, resource.CanView) {
			render403(request)
			return
		}

		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}

		var req listRequest
		err = json.Unmarshal(data, &req)
		if err != nil {
			panic(err)
		}

		listData, err := resource.getListContent(admin, &req, user)
		if err != nil {
			panic(err)
		}

		request.Response().Header().Set("X-Count", fmt.Sprintf("%d", listData.Count))
		request.Response().Header().Set("X-Total-Count", fmt.Sprintf("%d", listData.TotalCount))
		request.SetData("admin_list", listData)
		request.RenderView("admin_list_cells")
	})
}

func bindListResourceAPI(admin *Administration) {
	admin.AdminController.Get(admin.GetURL("_api/resource/:name"), func(request prago.Request) {
		locale := GetLocale(request)
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := admin.resourceNameMap[name]
		if !found {
			render404(request)
			return
		}

		if !admin.Authorize(*user, resource.CanView) {
			render403(request)
			return
		}

		var item interface{}
		resource.newItem(&item)
		c, err := admin.Query().Count(item)
		must(err)
		if c == 0 {
			request.RenderJSON([]string{})
			return
		}

		ret := []resourceItem{}

		var items interface{}
		resource.newArrayOfItems(&items)
		must(admin.Query().Get(items))

		itemsVal := reflect.ValueOf(items).Elem()

		for i := 0; i < itemsVal.Len(); i++ {
			item := itemsVal.Index(i)

			id := item.Elem().FieldByName("ID").Int()

			var name string
			ifaceItemName, ok := item.Interface().(interface {
				AdminItemName(string) string
			})
			if ok {
				name = ifaceItemName.AdminItemName(locale)
			} else {
				name = item.Elem().FieldByName("Name").String()
			}

			ret = append(ret, resourceItem{
				ID:   id,
				Name: name,
			})
		}
		request.RenderJSON(ret)
	})
}

func bindListResourceItemAPI(admin *Administration) {
	admin.AdminController.Get(admin.GetURL("_api/resource/:name/:id"), func(request prago.Request) {
		user := GetUser(request)
		resourceName := request.Params().Get("name")
		resource, found := admin.resourceNameMap[resourceName]
		if !found {
			render404(request)
			return
		}

		if !admin.Authorize(*user, resource.CanView) {
			render403(request)
			return
		}

		idStr := request.Params().Get("id")

		var item interface{}
		resource.newItem(&item)
		must(admin.Query().WhereIs("id", idStr).Get(item))

		ret := resourceItem{}

		itemVal := reflect.ValueOf(item).Elem()

		id := itemVal.FieldByName("ID").Int()

		var name string
		name = itemVal.FieldByName("Name").String()
		ret.ID = id
		ret.Name = name

		request.RenderJSON(ret)
	})
}
