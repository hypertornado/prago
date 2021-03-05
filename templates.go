package prago

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago/messages"
)

type templates struct {
	templates *template.Template
	funcMap   template.FuncMap
}

func (app *App) initTemplates() {
	app.templates = &templates{
		templates: template.New(""),
		funcMap:   map[string]interface{}{},
	}

	app.AddTemplateFunction("HTML", func(data string) template.HTML {
		return template.HTML(data)
	})

	app.AddTemplateFunction("HTMLAttr", func(data string) template.HTMLAttr {
		return template.HTMLAttr(data)
	})

	app.AddTemplateFunction("CSS", func(data string) template.CSS {
		return template.CSS(data)
	})

	app.AddTemplateFunction("tmpl", func(templateName string, x interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := app.templates.templates.ExecuteTemplate(&buf, templateName, x)
		return template.HTML(buf.String()), err
	})

	app.AddTemplateFunction("markdown", func(text string) template.HTML {
		return template.HTML(markdown.New(markdown.Breaks(true)).RenderToString([]byte(text)))
	})

	app.AddTemplateFunction("message", func(language, id string) template.HTML {
		return template.HTML(messages.Messages.Get(language, id))
	})

	app.AddTemplateFunction("thumb", func(ids string) string {
		return app.thumb(ids)
	})

	app.AddTemplateFunction("img", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			var image File
			err := app.Query().WhereIs("uid", v).Get(&image)
			if err == nil && image.IsImage() {
				return image.GetLarge()
			}
		}
		return ""
	})

	app.AddTemplateFunction("istabvisible", isTabVisible)

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

//LoadTemplateFromFS loads app's html templates from file system
func (app *App) LoadTemplateFromFS(fsys fs.FS, patterns ...string) (err error) {
	app.templates.templates, err = app.templates.templates.Funcs(app.templates.funcMap).ParseFS(fsys, patterns...)
	return
}

//AddTemplateFunction adds template function
func (app *App) AddTemplateFunction(name string, f interface{}) {
	app.templates.funcMap[name] = f
}

//ExecuteTemplate executes template
func (app *App) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return app.templates.templates.ExecuteTemplate(wr, name, data)
}
