package extensions

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/jinzhu/gorm"
	"html/template"
	"strconv"
)

type Admin struct {
	Prefix    string
	AppName   string
	Resources []*AdminResource
	db        *sql.DB
	gorm      *gorm.DB
}

func (a *Admin) Migrate() (err error) {
	for _, v := range a.Resources {
		err = v.Migrate()
		if err != nil {
			return
		}
	}
	return
}

func (a *Admin) AddResource(resource *AdminResource) {
	resource.admin = a
	a.Resources = append(a.Resources, resource)
}

func (a *Admin) adminHeaderData() interface{} {
	ret := map[string]interface{}{
		"appName": a.AppName,
		"prefix":  a.Prefix,
	}
	menuitems := []map[string]interface{}{}
	for _, resource := range a.Resources {
		menuitems = append(menuitems, map[string]interface{}{
			"name": resource.Name,
			"id":   resource.ID,
			"url":  a.Prefix + "/" + resource.ID,
		})
	}
	ret["menu"] = menuitems
	return ret
}

func (a *Admin) Init(app *prago.App) error {
	a.AppName = app.Data()["appName"].(string)
	a.db = app.Data()["db"].(*sql.DB)
	a.gorm = app.Data()["gorm"].(*gorm.DB)
	t := app.Data()["templates"].(*template.Template)
	if t == nil {
		panic("templates not found")
	}
	path := "/Users/ondrejodchazel/projects/go/src/github.com/hypertornado/prago/extensions/templates/"

	var err error
	err = a.Migrate()
	if err != nil {
		panic(err)
	}

	funcs := app.Data()["templateFuncs"].(template.FuncMap)
	if funcs == nil {
		panic("funcs not found")
	}

	funcs["tmpl"] = func(templateName string, x interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := t.ExecuteTemplate(&buf, templateName, x)
		return template.HTML(buf.String()), err
	}

	t = t.Funcs(funcs)
	t, err = t.ParseGlob(path + "/*.tmpl")
	if err != nil {
		panic(err)
	}

	app.Data()["templates"] = t

	adminController := app.MainController().SubController()

	adminController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("appName", request.App().Data()["appName"].(string))
		request.SetData("admin_header", a.adminHeaderData())

		request.SetData("admin_yield", "admin_home")

		next()
	})

	adminController.Get(a.Prefix, func(request prago.Request) {
		prago.Render(request, 200, "admin_layout")
	})

	type RowItem struct {
		Name string
		Url  string
	}

	for i, _ := range a.Resources {
		resource := a.Resources[i]

		resourceController := adminController.SubController()

		resourceController.AddBeforeAction(func(request prago.Request) {
			request.SetData("admin_resource", resource)
		})

		resourceController.Get(a.Prefix+"/"+resource.ID, func(request prago.Request) {
			row_items, err := resource.List()
			if err != nil {
				panic(err)
			}

			request.SetData("admin_items", row_items)
			request.SetData("admin_yield", "admin_list")
			prago.Render(request, 200, "admin_layout")
		})

		resourceController.Get(a.Prefix+"/"+resource.ID+"/new", func(request prago.Request) {

			//request.SetData("admin_node")
			request.SetData("admin_yield", "admin_new")
			prago.Render(request, 200, "admin_layout")
		})

		resourceController.Post(a.Prefix+"/"+resource.ID+"/new", func(request prago.Request) {
			fmt.Println(request.Params())
			prago.Redirect(request, a.Prefix+"/"+resource.ID)
		})

		resourceController.Get(a.Prefix+"/"+resource.ID+"/:id", func(request prago.Request) {

			id, err := strconv.Atoi(request.Params().Get("id"))
			if err != nil {
				panic(err)
			}

			fmt.Println(id)

			request.SetData("admin_yield", "admin_edit")
			prago.Render(request, 200, "admin_layout")
		})
	}

	return nil
}

func NewAdmin(prefix string) *Admin {
	return &Admin{
		Prefix:    prefix,
		Resources: []*AdminResource{},
	}
}
