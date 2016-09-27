package admin

import (
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"reflect"
)

func bindMarkdownAPI(a *Admin) {
	a.AdminController.Post(a.Prefix+"/_api/markdown", func(request prago.Request) {
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		prago.WriteAPI(request, markdown.New().RenderToString(data), 200)
	})
}

type resourceItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func bindListResourceAPI(a *Admin) {
	a.AdminController.Get(a.Prefix+"/_api/resource/:name", func(request prago.Request) {
		locale := GetLocale(request)
		user := GetUser(request)
		name := request.Params().Get("name")
		resource := a.resourceNameMap[name]

		if !resource.Authenticate(user) {
			render403(request)
			return
		}

		c, err := resource.Query().Count()
		prago.Must(err)
		if c == 0 {
			prago.WriteAPI(request, []string{}, 200)
			return
		}

		ret := []resourceItem{}

		items, err := resource.Query().List()
		prago.Must(err)

		itemsVal := reflect.ValueOf(items)

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
