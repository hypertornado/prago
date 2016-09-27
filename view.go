package prago

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
)

//Render outputs view of viewName with statusCode to request
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

type middlewareView struct{}

func (m middlewareView) Init(app *App) error {
	app.requestMiddlewares = append(app.requestMiddlewares, requestMiddlewareView)

	templates := template.New("")
	templateFuncs := template.FuncMap{}

	templateFuncs["Plain"] = func(data string) template.HTML {
		return template.HTML(data)
	}

	templateFuncs["CSS"] = func(data string) template.CSS {
		return template.CSS(data)
	}

	templates = templates.Funcs(templateFuncs)

	app.data["templates"] = templates
	app.data["templateFuncs"] = templateFuncs
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
		statusCode = http.StatusOK
		renderDefaultNotFound = true
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

//GetTemplates return app's html templates
func (app *App) GetTemplates() (*template.Template, template.FuncMap, error) {
	templates, ok := app.data["templates"].(*template.Template)
	if !ok {
		return nil, nil, errors.New("Templates not initialized")
	}

	templateFuncs, ok := app.data["templateFuncs"].(template.FuncMap)
	if !ok {
		return nil, nil, errors.New("Template function maps not initialized")
	}
	return templates, templateFuncs, nil
}

//LoadTemplatePath loads app's html templates from path pattern
func (app *App) LoadTemplatePath(pattern string) (err error) {
	templates, templateFuncs, err := app.GetTemplates()
	if err != nil {
		return err
	}

	templates = templates.Funcs(templateFuncs)
	templates, err = templates.ParseGlob(pattern)
	if err != nil {
		return err
	}

	app.data["templates"] = templates
	return nil
}

//LoadTemplateFromString loads app's html templates from string
func (app *App) LoadTemplateFromString(in string) (err error) {
	templates, templateFuncs, err := app.GetTemplates()
	if err != nil {
		return err
	}

	templates = templates.Funcs(templateFuncs)
	templates, err = templates.Parse(in)
	if err != nil {
		return err
	}

	app.data["templates"] = templates
	return nil

}

//AddTemplateFunction adds template function
func (app *App) AddTemplateFunction(name string, f interface{}) (err error) {
	_, templateFuncs, err := app.GetTemplates()
	if err != nil {
		return err
	}

	templateFuncs[name] = f
	app.data["templateFuncs"] = templateFuncs
	return nil
}
