package admin

import (
	"encoding/json"
	"fmt"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"reflect"
	"strings"
)

func bindAPI(a *Admin) {
	bindMarkdownAPI(a)
	bindListAPI(a)
	bindListResourceAPI(a)
	bindListResourceItemAPI(a)
}

func bindImageAPI(admin *Admin, fileDownloadPath string) {
	admin.AdminController.Get(admin.GetURL("file/uuid/:uuid"), func(request prago.Request) {
		var image File
		err := admin.Query().WhereIs("uid", request.Params().Get("uuid")).Get(&image)
		if err != nil {
			panic(err)
		}
		prago.Redirect(request,
			fmt.Sprintf("%s/file/%d", admin.Prefix, image.ID),
		)
	})

	admin.AdminController.Get(admin.GetURL("_api/image/thumb/:id"), func(request prago.Request) {
		var image File
		prago.Must(admin.Query().WhereIs("uid", request.Params().Get("id")).Get(&image))
		prago.Redirect(request, image.GetMedium())
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
			prago.Must(q.Get(&images))
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
			prago.Must(admin.Create(file))
			files = append(files, file)
		}

		writeFileResponse(request, files)
	})
}

func bindMarkdownAPI(admin *Admin) {
	admin.AdminController.Post(admin.GetURL("_api/markdown"), func(request prago.Request) {
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		prago.WriteAPI(request, markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data), 200)
	})
}

type resourceItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func bindListAPI(admin *Admin) {
	admin.AdminController.Post(admin.GetURL("_api/list/:name"), func(request prago.Request) {
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := admin.resourceNameMap[name]
		if !found {
			render404(request)
			return
		}

		if !resource.Authenticate(user) {
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
		prago.Render(request, 200, "admin_list_cells")
	})
}

func bindListResourceAPI(admin *Admin) {
	admin.AdminController.Get(admin.GetURL("_api/resource/:name"), func(request prago.Request) {
		locale := GetLocale(request)
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := admin.resourceNameMap[name]
		if !found {
			render404(request)
			return
		}

		if !resource.Authenticate(user) {
			render403(request)
			return
		}

		var item interface{}
		resource.newItem(&item)
		c, err := admin.Query().Count(item)
		prago.Must(err)
		if c == 0 {
			prago.WriteAPI(request, []string{}, 200)
			return
		}

		ret := []resourceItem{}

		var items interface{}
		resource.newItems(&items)
		prago.Must(admin.Query().Get(items))

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

		prago.WriteAPI(request, ret, 200)
	})
}

func bindListResourceItemAPI(admin *Admin) {
	admin.AdminController.Get(admin.GetURL("_api/resource/:name/:id"), func(request prago.Request) {
		user := GetUser(request)
		resourceName := request.Params().Get("name")
		resource, found := admin.resourceNameMap[resourceName]
		if !found {
			render404(request)
			return
		}

		if !resource.Authenticate(user) {
			render403(request)
			return
		}

		idStr := request.Params().Get("id")

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", idStr).Get(item))

		ret := resourceItem{}

		itemVal := reflect.ValueOf(item).Elem()

		id := itemVal.FieldByName("ID").Int()

		var name string
		name = itemVal.FieldByName("Name").String()

		ret.ID = id
		ret.Name = name

		prago.WriteAPI(request, ret, 200)
	})
}
