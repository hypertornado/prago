package admin

import (
	"encoding/json"
	"fmt"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"reflect"
)

func bindAPI(a *Admin) {
	bindMarkdownAPI(a)
	bindListAPI(a)
	bindListResourceAPI(a)
	bindListResourceItemAPI(a)
}

func bindMarkdownAPI(a *Admin) {
	a.AdminController.Post(a.Prefix+"/_api/markdown", func(request prago.Request) {
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

func bindListAPI(a *Admin) {
	a.AdminController.Post(a.Prefix+"/_api/list/:name", func(request prago.Request) {
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := a.resourceNameMap[name]
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

		listData, err := resource.getListContent(a, &req, user)
		if err != nil {
			panic(err)
		}

		request.Response().Header().Set("X-Count", fmt.Sprintf("%d", listData.Count))
		request.Response().Header().Set("X-Total-Count", fmt.Sprintf("%d", listData.TotalCount))
		request.SetData("admin_list", listData)
		prago.Render(request, 200, "admin_list_cells")
	})
}

func bindListResourceAPI(a *Admin) {
	a.AdminController.Get(a.Prefix+"/_api/resource/:name", func(request prago.Request) {
		locale := GetLocale(request)
		user := GetUser(request)
		name := request.Params().Get("name")
		resource, found := a.resourceNameMap[name]
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
		c, err := a.Query().Count(item)
		prago.Must(err)
		if c == 0 {
			prago.WriteAPI(request, []string{}, 200)
			return
		}

		ret := []resourceItem{}

		var items interface{}
		resource.newItems(&items)
		prago.Must(a.Query().Get(items))

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

func bindListResourceItemAPI(a *Admin) {
	a.AdminController.Get(a.Prefix+"/_api/resource/:name/:id", func(request prago.Request) {
		user := GetUser(request)
		resourceName := request.Params().Get("name")
		resource, found := a.resourceNameMap[resourceName]
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
		prago.Must(a.Query().WhereIs("id", idStr).Get(item))

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
