package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/utils"
	"github.com/sendgrid/sendgrid-go"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	ErrorNotFound = errors.New("ErrorNotFound")
)

type Admin struct {
	Prefix                string
	AppName               string
	Resources             []*Resource
	resourceMap           map[reflect.Type]*Resource
	resourceNameMap       map[string]*Resource
	AdminController       *prago.Controller
	AdminAccessController *prago.Controller
	App                   *prago.App
	db                    *sql.DB
	authData              map[string]string
	seedFn                func(*prago.App) error
	sendgridClient        *sendgrid.SGClient
	noReplyEmail          string
}

func NewAdmin(prefix, name string) *Admin {
	ret := &Admin{
		Prefix:          prefix,
		AppName:         name,
		Resources:       []*Resource{},
		resourceMap:     make(map[reflect.Type]*Resource),
		resourceNameMap: make(map[string]*Resource),
	}

	ret.CreateResource(User{})
	ret.CreateResource(File{})

	return ret
}

func (a *Admin) Seed(fn func(*prago.App) error) {
	a.seedFn = fn
}

func (a *Admin) CreateResource(item interface{}) (resource *Resource, err error) {
	resource, err = NewResource(item)
	if err != nil {
		return
	}
	err = a.AddResource(resource)
	return
}

func (a *Admin) UnsafeDropTables() error {
	for _, resource := range a.Resources {
		if resource.HasModel {
			err := resource.UnsafeDropTable()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Admin) Migrate(verbose bool) error {
	for _, resource := range a.Resources {
		if resource.HasModel {
			err := resource.migrate(verbose)
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

func (a *Admin) AddResource(resource *Resource) error {
	resource.admin = a
	a.Resources = append(a.Resources, resource)
	if resource.HasModel {
		a.resourceMap[resource.Typ] = resource
		a.resourceNameMap[resource.ID] = resource
	}
	return nil
}

func (a *Admin) GetResourceByName(name string) *Resource {
	return a.resourceNameMap[utils.ColumnName(name)]
}

func GetUser(request prago.Request) *User {
	return request.GetData("currentuser").(*User)
}

func (a *Admin) adminHeaderData(request prago.Request) interface{} {
	ret := map[string]interface{}{
		"appName": a.AppName,
		"prefix":  a.Prefix,
	}

	user := GetUser(request)
	locale := GetLocale(request)

	menuitems := []map[string]interface{}{}
	for _, resource := range a.Resources {
		newItem := map[string]interface{}{
			"name": resource.Name(locale),
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
	a.App = app
	a.db = app.Data()["db"].(*sql.DB)

	a.AdminAccessController = app.MainController().SubController()
	a.AdminController = a.AdminAccessController.SubController()

	bindDBBackupCron(app)
	BindMarkdownAPI(a)
	BindListResourceAPI(a)

	var err error

	a.sendgridClient = sendgrid.NewSendGridClientWithApiKey(app.Config().GetString("sendgridApi"))
	a.noReplyEmail = app.Config().GetString("noReplyEmail")

	err = a.bindAdminCommand(app)
	if err != nil {
		return err
	}

	err = a.initTemplates(app)
	if err != nil {
		return err
	}

	err = app.LoadTemplateFromString(TEMPLATES)
	if err != nil {
		panic(err)
	}

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

		randomness := app.Config().GetString("random")
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", GetLocale(request))

		request.SetData("appName", a.AppName)
		request.SetData("appCode", request.App().Data()["appName"].(string))
		request.SetData("appVersion", request.App().Data()["version"].(string))
		request.SetData("admin_header", a.adminHeaderData(request))

		next()
	})

	a.AdminController.Get(a.Prefix, func(request prago.Request) {
		prago.Render(request, 200, "admin_layout")
	})

	a.AdminController.Get(a.Prefix+"/_stats", stats)

	for i, _ := range a.Resources {
		resource := a.Resources[i]
		err = a.initResource(resource)
		if err != nil {
			return err
		}
	}

	a.AdminController.Get(a.Prefix+"/*", func(request prago.Request) {
		Render404(request)
	})

	return nil
}

func (a *Admin) bindAdminCommand(app *prago.App) error {
	adminCommand := app.CreateCommand("admin", "Admin tasks (migrate|seed|drop|thumbnails)")

	adminSubcommand := adminCommand.Arg("admincommand", "").Required().String()

	app.AddCommand(adminCommand, func(app *prago.App) error {
		switch *adminSubcommand {
		case "migrate":
			println("Migrating database")
			err := a.Migrate(true)
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
		case "thumbnails":
			println("Updating thumbnails")
			return UpdateFiles(a)
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
		return template.HTML(markdown.New(markdown.Breaks(true)).RenderToString([]byte(text)))
	})

	app.AddTemplateFunction("message", func(language, id string) template.HTML {
		return template.HTML(messages.Messages.Get(language, id))
	})

	app.AddTemplateFunction("thumb", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			var image File
			err := a.Query().WhereIs("uid", v).Get(&image)
			if err == nil && image.IsImage() {
				return image.GetSmall()
			}
		}
		return ""
	})

	return nil
}

func (a *Admin) GetURL(resource *Resource, suffix string) string {
	ret := a.Prefix + "/" + resource.ID
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

func bindDBBackupCron(app *prago.App) {
	app.AddCronTask("backup db", func() {
		err := extensions.BackupApp(app)
		if err != nil {
			app.Log().Error("Error while creating backup:", err)
		}
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
