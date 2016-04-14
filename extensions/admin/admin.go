package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
	"html/template"
	"os"
	"reflect"
	"strconv"
)

var (
	ErrorNotFound = errors.New("ErrorNotFound")
)

type Admin struct {
	Prefix          string
	AppName         string
	Resources       []*AdminResource
	resourceMap     map[reflect.Type]*AdminResource
	AdminController *prago.Controller
	db              *sql.DB
	authData        map[string]string
	seedFn          func(*prago.App) error
}

func NewAdmin(prefix, name string) *Admin {
	return &Admin{
		Prefix:      prefix,
		AppName:     name,
		Resources:   []*AdminResource{},
		resourceMap: make(map[reflect.Type]*AdminResource),
	}
}

func (a *Admin) Seed(fn func(*prago.App) error) {
	a.seedFn = fn
}

func (a *Admin) CreateResources(items ...interface{}) error {
	for _, item := range items {
		resource, err := NewResource(item)
		if err != nil {
			return err
		}
		err = a.AddResource(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Admin) SetAuthData(authData map[string]string) {
	a.authData = authData
}

func (a *Admin) UnsafeDropTables() error {
	for _, resource := range a.Resources {
		if resource.hasModel {
			err := resource.UnsafeDropTable()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Admin) Migrate() error {
	for _, resource := range a.Resources {
		if resource.hasModel {
			err := resource.Migrate()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Admin) AddResource(resource *AdminResource) error {
	resource.admin = a
	a.Resources = append(a.Resources, resource)
	if resource.hasModel {
		a.resourceMap[resource.Typ] = resource
	}
	return nil
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

	var err error

	err = a.bindAdminCommand(app)
	if err != nil {
		return err
	}

	err = a.initTemplates(app)
	if err != nil {
		return err
	}

	appName := app.Data()["appName"].(string)
	path := os.Getenv("HOME") + "/." + appName + "/files"
	BindImageResizer(app.MainController(), path)

	err = app.LoadTemplateFromString(TEMPLATES)
	if err != nil {
		panic(err)
	}

	adminAccessController := app.MainController().SubController()

	adminAccessController.Get(a.Prefix+"/login", func(request prago.Request) {
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("name", a.AppName)
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

	adminAccessController.Get(a.Prefix+"/admin.css", func(request prago.Request) {
		request.Response().Header().Add("Content-type", "text/css")
		request.SetData("statusCode", 200)
		request.SetData("body", []byte(CSS))
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

func (a *Admin) bindAdminCommand(app *prago.App) error {
	adminCommand := app.CreateCommand("admin", "Admin tasks")

	adminSubcommand := adminCommand.Arg("admincommand", "").Required().String()

	app.AddCommand(adminCommand, func(app *prago.App) error {
		switch *adminSubcommand {
		case "migrate":
			println("Migrating database")
			err := a.Migrate()
			if err == nil {
				println("Migrate done")
			}
			return err
		case "seed":
			if a.seedFn != nil {
				println("Seeding")
				return a.seedFn(app)
			} else {
				return errors.New("No seed function defined")
			}
		case "drop":
			if utils.ConsoleQuestion("Really want to drop table?") {
				println("Dropping table")
				return a.UnsafeDropTables()
			} else {
				return nil
			}
		default:
			println("unknown admin subcommand " + *adminSubcommand)
		}
		return nil
	})

	return nil
}

func (a *Admin) initTemplates(app *prago.App) error {
	templates := app.Data()["templates"].(*template.Template)
	if templates == nil {
		return errors.New("Templates not initialized")
	}

	app.AddTemplateFunction("tmpl", func(templateName string, x interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := templates.ExecuteTemplate(&buf, templateName, x)
		return template.HTML(buf.String()), err
	})

	app.AddTemplateFunction("markdown", func(text string) template.HTML {
		return template.HTML(markdown.New().RenderToString([]byte(text)))
	})

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

		tableData, err := resource.ListTableItems()
		if err != nil {
			panic(err)
		}

		request.SetData("admin_list_table_data", tableData)
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
		resource.adminStructCache.BindData(item, request.Params(), request.Request().MultipartForm, BindDataFilterDefault)
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

		err = resource.adminStructCache.BindData(item, request.Params(), request.Request().MultipartForm, BindDataFilterDefault)
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
