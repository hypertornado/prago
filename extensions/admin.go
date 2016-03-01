package extensions

import (
	"bytes"
	"database/sql"
	"errors"
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

func (a *Admin) DB() *sql.DB {
	return a.db
}

func (a *Admin) Init(app *prago.App) error {
	a.db = app.Data()["db"].(*sql.DB)
	a.gorm = app.Data()["gorm"].(*gorm.DB)

	err := a.initTemplates(app)
	if err != nil {
		return err
	}

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

func (a *Admin) initTemplates(app *prago.App) error {
	templates := app.Data()["templates"].(*template.Template)
	if templates == nil {
		return errors.New("Templates not initialized")
	}

	templateFuncs := app.Data()["templateFuncs"].(template.FuncMap)
	if templateFuncs == nil {
		return errors.New("Funcs not initialized")
	}

	templateFuncs["tmpl"] = func(templateName string, x interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := templates.ExecuteTemplate(&buf, templateName, x)
		return template.HTML(buf.String()), err
	}

	templates = templates.Funcs(templateFuncs)

	app.Data()["templates"] = templates
	app.Data()["templateFuncs"] = templateFuncs

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

func (a *Admin) GetURL(resource *AdminResource, suffix string) string {
	ret := a.Prefix + "/" + resource.ID
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

func BindList(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ""), func(request prago.Request) {

		listTableHeader, err := resource.ListTableHeader()
		if err != nil {
			panic(err)
		}

		/*q := resource.Query()
		q = resource.queryFilter(q)
		row_items, err := q.List()
		if err != nil {
			panic(err)
		}*/

		listTableItems, err := resource.ListTableItems()
		if err != nil {
			panic(err)
		}

		request.SetData("admin_list_table_items", listTableItems)
		request.SetData("admin_list_table_header", listTableHeader)
		//request.SetData("admin_items", row_items)
		request.SetData("admin_yield", "admin_list")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindNew(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, "new"), func(request prago.Request) {

		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}

		formItems, err := resource.GetFormItems(item)
		if err != nil {
			panic(err)
		}

		request.SetData("admin_form_items", formItems)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindCreate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}
		BindData(item, request.Params(), BindDataFilterDefault)
		err = resource.Create(item)
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDetail(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		if err != nil {
			panic(err)
		}

		formItems, err := resource.GetFormItems(item)
		if err != nil {
			panic(err)
		}

		request.SetData("admin_item", item)
		request.SetData("admin_form_items", formItems)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindUpdate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		if err != nil {
			panic(err)
		}

		err = BindData(item, request.Params(), BindDataFilterDefault)
		if err != nil {
			panic(err)
		}

		err = resource.Save(item)
		if err != nil {
			panic(err)
		}

		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDelete(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id/delete"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		_, err = resource.Query().Where(map[string]interface{}{"id": int64(id)}).Delete()
		if err != nil {
			panic(err)
		}

		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func AdminInitResourceDefault(a *Admin, resource *AdminResource) error {
	BindList(a, resource)
	BindNew(a, resource)
	BindCreate(a, resource)
	BindDetail(a, resource)
	BindUpdate(a, resource)
	BindDelete(a, resource)
	return nil
}

func NewAdmin(prefix, name string) *Admin {
	return &Admin{
		Prefix:    prefix,
		AppName:   name,
		Resources: []*AdminResource{},
	}
}
