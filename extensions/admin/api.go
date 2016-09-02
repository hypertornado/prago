package admin

import (
	"encoding/json"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"reflect"
)

func BindMarkdownAPI(a *Admin) {
	a.AdminController.Post(a.Prefix+"/_api/markdown", func(request prago.Request) {
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		WriteApi(request, markdown.New().RenderToString(data), 200)
	})
}

type resourceItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func BindListResourceAPI(a *Admin) {
	a.AdminController.Get(a.Prefix+"/_api/resource/:name", func(request prago.Request) {
		locale := GetLocale(request)
		user := GetUser(request)
		name := request.Params().Get("name")
		resource := a.resourceNameMap[name]

		if !resource.Authenticate(user) {
			panic("EEE")
		}

		c, err := resource.Query().Count()
		prago.Must(err)
		if c == 0 {
			WriteApi(request, []string{}, 200)
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

		WriteApi(request, ret, 200)
	})
}

func WriteApi(r prago.Request, data interface{}, code int) {
	r.SetProcessed()

	r.Response().Header().Add("Content-type", "application/json")

	pretty := false
	if r.Params().Get("pretty") == "true" {
		pretty = true
	}

	var responseToWrite interface{}
	if code >= 400 {
		responseToWrite = map[string]interface{}{"error": data, "errorCode": code}
	} else {
		responseToWrite = data
	}

	var result []byte
	var e error

	if pretty == true {
		result, e = json.MarshalIndent(responseToWrite, "", "  ")
	} else {
		result, e = json.Marshal(responseToWrite)
	}

	if e != nil {
		panic("error while generating JSON output")
	}
	r.Response().WriteHeader(code)
	r.Response().Write(result)
}
