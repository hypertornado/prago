package prago

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"sync"

	"github.com/golang-commonmark/markdown"
)

//TODO: use https://github.com/yuin/goldmark

//go:embed templates
var templatesFS embed.FS

type PragoTemplates struct {
	templates      *template.Template
	funcMap        template.FuncMap
	templatesMutex *sync.RWMutex
	fs             fs.FS
	matchPatterns  []string
	watchPattern   string
}

func NewPragoTemplates() *PragoTemplates {
	return &PragoTemplates{
		funcMap:        map[string]interface{}{},
		templatesMutex: &sync.RWMutex{},
	}
}

func (app *App) initTemplates() {
	app.adminTemplates = NewPragoTemplates()

	app.adminTemplates.Function("PragoCSS", func(data string) template.CSS {
		return template.CSS(data)
	})

	app.adminTemplates.Function("PragoMarkdown", func(text string) template.HTML {
		return template.HTML(markdown.New(markdown.Breaks(true)).RenderToString([]byte(text)))
	})

	app.adminTemplates.Function("PragoMessage", func(language, id string) template.HTML {
		return template.HTML(messages.Get(language, id))
	})

	app.adminTemplates.Function("PragoThumb", func(ids string) string {
		return app.thumb(ids)
	})

	app.adminTemplates.Function("PragoIconExists", func(iconName string) bool {
		return app.iconExists(iconName)
	})

	must(app.adminTemplates.SetFilesystem(templatesFS, "templates/*.tmpl"))
}

func (templates *PragoTemplates) SetFilesystem(fsys fs.FS, patterns ...string) error {
	templates.templatesMutex.Lock()
	defer templates.templatesMutex.Unlock()
	templates.fs = fsys
	templates.matchPatterns = patterns
	return templates.parseTemplates()
}

func (templates *PragoTemplates) parseTemplates() error {
	t := template.New("")
	t = t.Funcs(templates.funcMap)
	var err error
	t, err = t.ParseFS(templates.fs, templates.matchPatterns...)
	if err != nil {
		return err
	}
	templates.templates = t
	return nil
}

func (templates *PragoTemplates) Function(name string, f interface{}) {
	templates.templatesMutex.Lock()
	defer templates.templatesMutex.Unlock()
	templates.funcMap[name] = f
}

func (templates *PragoTemplates) watch(app *App) {
	if templates.watchPattern == "" {
		return
	}
	app.watchPath("tmpl", templates.watchPattern, func() {
		templates.templatesMutex.Lock()
		defer templates.templatesMutex.Unlock()
		err := templates.parseTemplates()
		if err != nil {
			app.Log().Printf("Error while compiling templates in development mode from path '%s': %s", templates.watchPattern, err)
		}
	})

}

func (templates *PragoTemplates) Execute(wr io.Writer, name string, data interface{}) error {
	templates.templatesMutex.RLock()
	defer templates.templatesMutex.RUnlock()
	return templates.templates.ExecuteTemplate(wr, name, data)
}

func (templates *PragoTemplates) ExecuteToString(templateName string, data interface{}) string {
	bufStats := new(bytes.Buffer)
	err := templates.Execute(bufStats, templateName, data)
	must(err)
	return bufStats.String()
}

func (templates *PragoTemplates) ExecuteToHTML(templateName string, data interface{}) template.HTML {
	return template.HTML(templates.ExecuteToString(templateName, data))
}
