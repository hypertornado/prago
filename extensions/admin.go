package extensions

//https://groups.google.com/forum/#!topic/golang-nuts/PRLloHrJrDU/discussion

import (
	"bytes"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/jinzhu/gorm"
	"html/template"
)

type Admin struct {
	Prefix    string
	AppName   string
	DB        gorm.DB
	Resources []*AdminResource
}

type AdminResource struct {
	ID    string
	Name  string
	Items []*AdminResourceItem
}

func (ar *AdminResource) AddItem(item *AdminResourceItem) {
	ar.Items = append(ar.Items, item)
}

func NewAdminInput(name string) *AdminResourceItem {
	return &AdminResourceItem{
		Name:     name,
		Template: "admin_input",
	}
}

type AdminResourceItem struct {
	Name     string
	Template string
}

func (a *Admin) AddResource(id, name string, resource interface{}) *AdminResource {
	ret := &AdminResource{
		ID:   id,
		Name: name,
	}
	a.Resources = append(a.Resources, ret)
	return ret
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

	t := app.Data()["templates"].(*template.Template)
	if t == nil {
		panic("templates not found")
	}
	path := "/Users/ondrejodchazel/projects/go/src/github.com/hypertornado/prago/extensions/templates/"

	var err error

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

	adminController.Get(a.Prefix, func(request prago.Request) {})

	for i, _ := range a.Resources {
		resource := a.Resources[i]

		resourceController := adminController.SubController()

		resourceController.AddBeforeAction(func(request prago.Request) {
			request.SetData("admin_resource", resource)
		})

		resourceController.Get(a.Prefix+"/"+resource.ID, func(request prago.Request) {
			request.SetData("admin_yield", "admin_list")
			prago.Render(request, 200, "admin_layout")
		})

		resourceController.Get(a.Prefix+"/"+resource.ID+"/new", func(request prago.Request) {

			request.SetData("admin_node", resource.Items)
			request.SetData("admin_yield", "admin_new")
			prago.Render(request, 200, "admin_layout")
		})

		resourceController.Post(a.Prefix+"/"+resource.ID+"/new", func(request prago.Request) {
			fmt.Println(request.Params())
			prago.Redirect(request, a.Prefix+"/"+resource.ID)
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
