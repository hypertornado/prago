package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/utils"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"time"
)

var (
	ErrorNotFound = errors.New("ErrorNotFound")
)

type Admin struct {
	Prefix                string
	AppName               string
	Resources             []*AdminResource
	resourceMap           map[reflect.Type]*AdminResource
	AdminController       *prago.Controller
	AdminAccessController *prago.Controller
	db                    *sql.DB
	authData              map[string]string
	seedFn                func(*prago.App) error
}

func NewAdmin(prefix, name string) *Admin {
	ret := &Admin{
		Prefix:      prefix,
		AppName:     name,
		Resources:   []*AdminResource{},
		resourceMap: make(map[reflect.Type]*AdminResource),
	}

	ret.CreateResources(User{})

	return ret
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

func FlashMessage(request prago.Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	prago.Must(session.Save(request.Request(), request.Response()))
}

func (a *Admin) AddResource(resource *AdminResource) error {
	resource.admin = a
	a.Resources = append(a.Resources, resource)
	if resource.hasModel {
		a.resourceMap[resource.Typ] = resource
	}
	return nil
}

func (a *Admin) GetUser(request prago.Request) *User {
	return request.GetData("currentuser").(*User)
}

func (a *Admin) adminHeaderData(request prago.Request) interface{} {
	ret := map[string]interface{}{
		"appName": a.AppName,
		"prefix":  a.Prefix,
	}

	user := a.GetUser(request)

	menuitems := []map[string]interface{}{}
	for _, resource := range a.Resources {
		newItem := map[string]interface{}{
			"name": resource.Name,
			"id":   resource.ID,
			"url":  a.Prefix + "/" + resource.ID,
		}

		if resource.Authenticate(user) {
			menuitems = append(menuitems, newItem)
		}
	}
	ret["menu"] = menuitems
	return ret
}

func (a *Admin) DB() *sql.DB {
	return a.db
}

func (a *Admin) Init(app *prago.App) error {
	a.db = app.Data()["db"].(*sql.DB)
	bindDBBackupCron(app)

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

	a.AdminAccessController = app.MainController().SubController()

	a.AdminController = a.AdminAccessController.SubController()

	a.AdminController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("admin_yield", "admin_home")

		session := request.GetData("session").(*sessions.Session)

		userId, ok := session.Values["user_id"].(int64)

		if !ok {
			prago.Redirect(request, a.Prefix+"/user/login")
			return
		}

		var user User
		err := a.Query().WhereIs("id", userId).Get(&user)
		if err != nil {
			prago.Redirect(request, a.Prefix+"/user/login")
			return

		}

		config, err := request.App().Config()
		prago.Must(err)
		randomness := config["random"]
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", GetLocale(request))

		request.SetData("appName", request.App().Data()["appName"].(string))
		request.SetData("admin_header", a.adminHeaderData(request))

		next()
	})

	a.AdminController.Get(a.Prefix, func(request prago.Request) {
		prago.Render(request, 200, "admin_layout")
	})

	a.AdminController.Get(a.Prefix+"/chunked", func(request prago.Request) {

		request.Response().WriteHeader(200)

		chunked := httputil.NewChunkedWriter(request.Response())

		for i := 0; i < 300; i++ {
			chunked.Write([]byte("xr\n"))
			time.Sleep(10 * time.Millisecond)
			request.Response().Header().Set("Content-Length", "0")
		}

		chunked.Close()

		request.SetProcessed()

	})

	a.AdminController.Get(a.Prefix+"/dump.sql", func(request prago.Request) {
		config, err := request.App().Config()
		if err != nil {
			panic(err)
		}

		user := config["dbUser"]
		dbName := config["dbName"]
		password := config["dbPassword"]

		cmd := exec.Command("mysqldump", "-u"+user, "-p"+password, dbName)

		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		request.Response().WriteHeader(200)

		var finished chan bool

		go func() {
			out, err := ioutil.ReadAll(outPipe)
			if err != nil {
				panic(err)
			}
			request.Response().Write(out)

			flusher, ok := request.Response().(http.Flusher)
			if !ok {
				panic(ok)
			}

			flusher.Flush()

			println("flushed")

			finished <- true
		}()

		err = cmd.Start()
		if err != nil {
			panic(err)
		}

		err = cmd.Wait()
		if err != nil {
			panic(err)
		}

		println("wait")

		<-finished

		request.Response().Header().Set("Content-Length", "0")

		println("finished")

		request.SetProcessed()
	})

	for i, _ := range a.Resources {
		resource := a.Resources[i]
		err = a.initResource(resource)
		if err != nil {
			return err
		}
	}

	a.AdminController.Get(a.Prefix+"/*", func(request prago.Request) {
		request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_404"))
		request.SetData("admin_yield", "admin_message")
		prago.Render(request, 200, "admin_layout")
	})

	return nil
}

func (a *Admin) bindAdminCommand(app *prago.App) error {
	adminCommand := app.CreateCommand("admin", "Admin tasks (migrate|seed|drop)")

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

	app.AddTemplateFunction("message", func(language, id string) template.HTML {
		return template.HTML(messages.Messages.Get(language, id))
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

	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		user := request.GetData("currentuser").(*User)
		if !resource.Authenticate(user) {
			request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_403"))
			request.SetData("admin_yield", "admin_message")
			prago.Render(request, 403, "admin_layout")
		} else {
			next()
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

		tableData, err := resource.ListTableItems(GetLocale(request))
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

		form, err := resource.StructCache.GetForm(item, GetLocale(request), DefaultVisibilityFilter, DefaultEditabilityFilter)
		if err != nil {
			panic(err)
		}

		form.Action = "../" + resource.ID
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
		AddCSRFToken(form, request)

		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindCreate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		ValidateCSRF(request)
		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}
		resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, BindDataFilterDefault)
		err = resource.Create(item)
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
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

		form, err := resource.StructCache.GetForm(item, GetLocale(request), DefaultVisibilityFilter, DefaultEditabilityFilter)
		if err != nil {
			panic(err)
		}

		form.Action = request.Params().Get("id")
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
		AddCSRFToken(form, request)

		request.SetData("admin_item", item)
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindUpdate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		if err != nil {
			panic(err)
		}

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, BindDataFilterDefault)
		if err != nil {
			panic(err)
		}

		fmt.Println(item)

		err = resource.Save(item)
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDelete(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id/delete"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		_, err = resource.Query().Where(map[string]interface{}{"id": int64(id)}).Delete()
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
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

func bindDBBackupCron(app *prago.App) {
	config, err := app.Config()
	if err != nil {
		panic(err)
	}

	user := config["dbUser"]
	dbName := config["dbName"]
	password := config["dbPassword"]

	app.AddCronTask("backup db", func() {
		app.Log().Println("Creating backup")
		cmd := exec.Command("mysqldump", "-u"+user, "-p"+password, dbName)

		dirPath := app.DotPath() + "/backups"
		os.Mkdir(dirPath, 0777)

		filePath := dirPath + "/" + time.Now().Format("2006_01_02_15_04_05") + ".sql"

		file, err := os.Create(filePath)
		if err != nil {
			app.Log().Error("Error while creating backup file:", err)
			return
		}

		cmd.Stdout = file
		defer file.Close()

		err = cmd.Run()
		if err != nil {
			app.Log().Error("Error while creating backup:", err)
			return
		}

		app.Log().Println("Backup created at:", filePath)

	}, func(t time.Time) time.Time {
		return t.AddDate(0, 0, 1)
	})

	app.AddCronTask("remove old backups", func() {
		app.Log().Println("Removing old backups")
		deadline := time.Now().AddDate(0, 0, -7)
		backupPath := app.DotPath() + "/backups"
		files, err := ioutil.ReadDir(backupPath)
		if err != nil {
			app.Log().Println("error while removing old backups:", err)
			return
		}

		for _, file := range files {
			if file.ModTime().Before(deadline) {
				removePath := backupPath + "/" + file.Name()
				err := os.Remove(removePath)
				if err != nil {
					app.Log().Println("Error while removing old backup file:", err)
				}
			}
		}
		app.Log().Println("Old backups removed")
	}, func(t time.Time) time.Time {
		return t.Add(1 * time.Hour)
	})

}
