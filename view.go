package prago

import (
	"bytes"
	"html/template"
	"net/http"
	"reflect"
)

func Render(request Request, statusCode int, viewName string) {
	buf := new(bytes.Buffer)
	request.Header().Add("Content-type", "text/html")
	request.Header().Add("Framework", "prago")

	request.SetData("statusCode", statusCode)

	err := request.App().Templates().ExecuteTemplate(buf, viewName, request.AllRequestData())
	if err != nil {
		panic(err)
	}

	request.SetData("body", buf.Bytes())

}

func MiddlewareWriteResponse(p Request) {
	if p.IsProcessed() {
		return
	}

	renderDefaultNotFound := false

	statusCode, statusCodeOk := p.GetData("statusCode").(int)
	if !statusCodeOk {
		if len(p.Header().Get("Location")) > 0 {
			statusCode = http.StatusMovedPermanently
		} else {
			statusCode = http.StatusOK
			renderDefaultNotFound = true
		}
	}
	body, bodyOk := p.GetData("body").([]byte)
	if !bodyOk {
		body = []byte{}
	} else {
		renderDefaultNotFound = false
	}

	if renderDefaultNotFound {
		statusCode = http.StatusNotFound
		body = []byte("404 - not found")
	}

	p.Response().WriteHeader(statusCode)
	p.Response().Write(body)
	p.SetProcessed()
}

func LoadTemplates(patterns []string, translations *I18N) (t *template.Template, err error) {
	funcs := template.FuncMap{}
	funcs["T"] = func(locale interface{}, id string) (string, error) {
		localeStr := ""
		if reflect.ValueOf(locale).Kind() == reflect.String {
			localeStr = reflect.ValueOf(locale).String()
		}
		return translations.GetTranslation(localeStr, id), nil
	}

	funcs["Plain"] = func(data string) template.HTML {
		return template.HTML(data)
	}

	t = template.New("")
	t = t.Funcs(funcs)
	for _, v := range patterns {
		t, err = t.ParseGlob(v)
		if err != nil {
			return nil, err
		}
	}
	return t, err
}
