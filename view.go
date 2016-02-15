package prago

import (
	"bytes"
	"html/template"
	"net/http"
	"reflect"
)

func Render(request Request, statusCode int, viewName string) {
	templates, ok := request.App().data["templates"].(*template.Template)
	if !ok {
		panic("couldnt find templates")
	}

	buf := new(bytes.Buffer)
	request.Header().Add("Content-type", "text/html")
	request.SetData("statusCode", statusCode)

	err := templates.ExecuteTemplate(buf, viewName, request.AllRequestData())
	if err != nil {
		panic(err)
	}

	request.SetData("body", buf.Bytes())
}

type MiddlewareView struct{}

func (m MiddlewareView) Init(app *App) error {
	translations := NewI18N()

	funcs := template.FuncMap{}

	templates, err := loadTemplates(
		funcs,
		[]string{
			"server/templates/*.tmpl",
		}, translations)
	if err != nil {
		panic(err)
	}

	app.data["templates"] = templates
	app.data["templateFuncs"] = funcs

	app.requestMiddlewares = append(app.requestMiddlewares, requestMiddlewareView)
	return nil
}

func requestMiddlewareView(p Request, next func()) {
	next()

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

func loadTemplates(funcs template.FuncMap, patterns []string, translations *I18N) (t *template.Template, err error) {
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
