package admin

import (
	"encoding/json"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"io/ioutil"
)

func BindMarkdownAPI(a *Admin) {
	a.App.MainController().Post(a.Prefix+"/_api/markdown", func(request prago.Request) {
		data, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}
		WriteApi(request, markdown.New().RenderToString(data), 200)
	})
}

func BindListResourceAPI(a *Admin) {
	a.App.MainController().Get(a.Prefix+"/_api/resource/:name", func(request prago.Request) {
		name := request.Params().Get("name")
		resource := a.resourceNameMap[name]
		c, err := resource.Query().Count()
		prago.Must(err)
		if c == 0 {
			WriteApi(request, []string{}, 200)
			return
		}
		items, err := resource.Query().List()
		prago.Must(err)
		WriteApi(request, items, 200)
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
