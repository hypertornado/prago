package prago

import (
	"bytes"
	"html/template"
)

type templates struct {
	templates *template.Template
	funcMap   template.FuncMap
}

func newTemplates() (ret *templates) {
	ret = &templates{
		templates: template.New(""),
		funcMap:   map[string]interface{}{},
	}

	ret.funcMap["HTML"] = func(data string) template.HTML {
		return template.HTML(data)
	}

	ret.funcMap["HTMLAttr"] = func(data string) template.HTMLAttr {
		return template.HTMLAttr(data)
	}

	ret.funcMap["CSS"] = func(data string) template.CSS {
		return template.CSS(data)
	}

	ret.funcMap["tmpl"] = func(templateName string, x interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := ret.templates.ExecuteTemplate(&buf, templateName, x)
		return template.HTML(buf.String()), err
	}

	return
}

//LoadTemplatePath loads app's html templates from path pattern
func (app *App) LoadTemplatePath(pattern string) (err error) {
	app.templates.templates, err = app.templates.templates.Funcs(
		app.templates.funcMap,
	).ParseGlob(pattern)
	return
}

//LoadTemplateFromString loads app's html templates from string
func (app *App) LoadTemplateFromString(in string) (err error) {
	app.templates.templates, err = app.templates.templates.Funcs(app.templates.funcMap).Parse(in)
	return
}

//AddTemplateFunction adds template function
func (app *App) AddTemplateFunction(name string, f interface{}) {
	app.templates.funcMap[name] = f
}
