package prago

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/golang-commonmark/markdown"
)

//TODO: use https://github.com/yuin/goldmark

//go:embed templates
var templatesFS embed.FS

type templates struct {
	templates      *template.Template
	funcMap        template.FuncMap
	templatesMutex *sync.RWMutex
	fileSystems    []*templateFS
}

type templateFS struct {
	fs       fs.FS
	patterns []string
}

func (app *App) initTemplates() {
	app.templates = &templates{
		funcMap:        map[string]interface{}{},
		templatesMutex: &sync.RWMutex{},
		fileSystems:    []*templateFS{},
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
		return template.HTML(messages.Get(language, id))
	})

	app.AddTemplateFunction("thumb", func(ids string) string {
		return app.thumb(context.TODO(), ids)
	})

	app.AddTemplateFunction("thumbnailExactSize", func(ids string, width, height int) string {
		return app.thumbnailExactSize(context.TODO(), ids, width, height)
	})

	app.AddTemplateFunction("img", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			image := Query[File](app).Is("uid", v).First()
			if image != nil && image.isImage() {
				return image.GetLarge()
			}
		}
		return ""
	})

	app.AddTemplateFunction("iconExists", func(iconName string) bool {
		return app.iconExists(iconName)
	})

	app.AddTemplateFunction("multiplication", func(a, b int) int { return a * b })

	must(app.AddTemplates(templatesFS, "templates/*.tmpl"))
}

// AddTemplates loads app's html templates from file system
func (app *App) AddTemplates(fsys fs.FS, patterns ...string) error {
	app.templates.templatesMutex.Lock()
	defer app.templates.templatesMutex.Unlock()

	tempFS := &templateFS{
		fs:       fsys,
		patterns: patterns,
	}

	app.templates.fileSystems = append(app.templates.fileSystems, tempFS)
	return app.parseTemplates()
}

func (app *App) parseTemplates() error {
	t := template.New("")
	t = t.Funcs(app.templates.funcMap)
	for _, v := range app.templates.fileSystems {
		var err error
		t, err = t.ParseFS(v.fs, v.patterns...)
		if err != nil {
			return err
		}
	}
	app.templates.templates = t
	return nil
}

// AddTemplateFunction adds template function
func (app *App) AddTemplateFunction(name string, f interface{}) {
	app.templates.templatesMutex.Lock()
	defer app.templates.templatesMutex.Unlock()
	app.templates.funcMap[name] = f
}

// ExecuteTemplate executes template
func (app *App) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	app.templates.templatesMutex.RLock()
	defer app.templates.templatesMutex.RUnlock()
	return app.templates.templates.ExecuteTemplate(wr, name, data)
}

// ExecuteTemplateToString executes template and return string, it panics
func (app *App) ExecuteTemplateToString(templateName string, data interface{}) string {
	bufStats := new(bytes.Buffer)
	err := app.ExecuteTemplate(bufStats, templateName, data)
	must(err)
	return bufStats.String()
}
