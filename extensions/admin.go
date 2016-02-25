package extensions

import (
	"bytes"
	"database/sql"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/jinzhu/gorm"
	"html/template"
	"strconv"
)

type Admin struct {
	Prefix          string
	AppName         string
	Resources       []*AdminResource
	AdminController *prago.Controller
	db              *sql.DB
	gorm            *gorm.DB
	authData        map[string]string
}

func (a *Admin) SetAuthData(authData map[string]string) {
	a.authData = authData
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
	a.db = app.Data()["db"].(*sql.DB)
	a.gorm = app.Data()["gorm"].(*gorm.DB)
	t := app.Data()["templates"].(*template.Template)
	if t == nil {
		panic("templates not found")
	}

	//TODO: read from config
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

	adminAccessController := app.MainController().SubController()

	adminAccessController.Get(a.Prefix+"/login", func(request prago.Request) {
		prago.Render(request, 200, "admin_login")
	})

	adminAccessController.Get(a.Prefix+"/logout", func(request prago.Request) {
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "email")
		err := session.Save(request.Request(), request.Response())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix+"/login")
	})

	adminAccessController.Post(a.Prefix+"/login", func(request prago.Request) {
		email := request.Params().Get("email")
		password := request.Params().Get("password")

		session := request.GetData("session").(*sessions.Session)
		requestedPassword, validUser := a.authData[email]
		if validUser && password == requestedPassword {
			session.Values["email"] = email
		} else {
			prago.Redirect(request, a.Prefix+"/login")
			return
		}

		err := session.Save(request.Request(), request.Response())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix)
	})

	a.AdminController = adminAccessController.SubController()

	a.AdminController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("appName", request.App().Data()["appName"].(string))
		request.SetData("admin_header", a.adminHeaderData())

		request.SetData("admin_yield", "admin_home")

		session := request.GetData("session").(*sessions.Session)

		email, ok := session.Values["email"].(string)
		_, userFound := a.authData[email]

		if !ok || !userFound {
			prago.Redirect(request, a.Prefix+"/login")
			return
		}

		request.SetData("admin_header_email", email)
		next()
	})

	a.AdminController.Get(a.Prefix, func(request prago.Request) {
		prago.Render(request, 200, "admin_layout")
	})

	for i, _ := range a.Resources {
		resource := a.Resources[i]
		err = a.initResource(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Admin) initResource(resource *AdminResource) error {

	resource.ResourceController = a.AdminController.SubController()

	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("admin_resource", resource)
		next()

		if !request.IsProcessed() && request.GetData("statusCode") == nil {
			prago.Render(request, 200, "admin_layout")
		}
	})

	init, ok := resource.item.(interface {
		AdminInitResource(*Admin, *AdminResource) error
	})

	if ok {
		return init.AdminInitResource(a, resource)
	} else {
		return AdminInitResourceDefault(a, resource)
	}
}

func AdminInitResourceDefault(a *Admin, resource *AdminResource) error {
	resourceController := resource.ResourceController

	resourceController.Get(resource.ResourceURL(""), func(request prago.Request) {
		row_items, err := resource.List()
		if err != nil {
			panic(err)
		}
		request.SetData("admin_items", row_items)
		request.SetData("admin_yield", "admin_list")
		prago.Render(request, 200, "admin_layout")
	})

	resourceController.Get(resource.ResourceURL("new"), func(request prago.Request) {

		descriptions, err := resource.GetItems()
		if err != nil {
			panic(err)
		}

		request.SetData("admin_item_descriptions", descriptions)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
	})

	resourceController.Post(resource.ResourceURL(""), func(request prago.Request) {
		err := resource.CreateItemFromParams(request.Params())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})

	resourceController.Get(resource.ResourceURL(":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		item, descriptions, err := resource.Get(int64(id))
		if err != nil {
			panic(err)
		}

		request.SetData("admin_item", item)
		request.SetData("admin_item_descriptions", descriptions)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
	})

	resourceController.Post(resource.ResourceURL(":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		err = resource.UpdateItemFromParams(int64(id), request.Params())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})

	resourceController.Post(resource.ResourceURL(":id/delete"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		err = resource.Delete(int64(id))
		if err != nil {
			panic(err)
		}

		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
	return nil
}

func NewAdmin(prefix, name string) *Admin {
	return &Admin{
		Prefix:    prefix,
		AppName:   name,
		Resources: []*AdminResource{},
	}
}
