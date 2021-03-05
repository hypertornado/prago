package prago

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	_ "embed"

	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

//go:embed static/public/admin/_static/admin.js
var staticAdminJS []byte

//go:embed static/public/admin/_static/admin.css
var staticAdminCSS []byte

//go:embed static/public/admin/_static/pikaday.js
var staticPikadayJS []byte

//go:embed templates
var templatesFS embed.FS

//Administration is struct representing admin extension
type AdministrationOLD struct {
	App              *App
	Logo             string
	prefix           string
	HumanName        string
	resources        []*Resource
	resourceMap      map[reflect.Type]*Resource
	resourceNameMap  map[string]*Resource
	accessController *Controller
	AdminController  *Controller
	rootActions      []Action
	db               *sql.DB
	resourcesInited  bool

	sendgridKey  string
	noReplyEmail string

	Newsletter *NewsletterMiddleware

	search *adminSearch

	fieldTypes  map[string]FieldType
	javascripts []string
	css         []string
	roles       map[string]map[string]bool

	activityListeners []func(ActivityLog)
	taskManager       *taskManager
}

//NewAdministration creates new administration on prefix url with name
func NewAdministration(app *App, initFunction func(*App)) *App {
	admin := app
	app.HumanName = app.AppName
	app.prefix = "/admin"
	app.resourceMap =
		make(map[reflect.Type]*Resource)
	app.resourceNameMap =
		make(map[string]*Resource)
	app.accessController =
		app.MainController().SubController()

	app.sendgridKey =
		app.Config.GetStringWithFallback("sendgridApi", "")
	app.noReplyEmail =
		app.Config.GetStringWithFallback("noReplyEmail", "")

	app.fieldTypes = make(map[string]FieldType)
	app.roles = make(map[string]map[string]bool)

	admin.taskManager = newTaskManager(admin)

	db, err := connectMysql(
		app.Config.GetStringWithFallback("dbUser", ""),
		app.Config.GetStringWithFallback("dbPassword", ""),
		app.Config.GetStringWithFallback("dbName", ""),
	)
	if err != nil {
		panic(err)
	}
	admin.db = db

	admin.AdminController = admin.accessController.SubController()
	admin.addDefaultFieldTypes()

	admin.CreateResource(User{}, initUserResource)
	admin.CreateResource(Notification{}, initNotificationResource)
	admin.CreateResource(File{}, initFilesResource)
	admin.CreateResource(ActivityLog{}, initActivityLog)

	if initFunction != nil {
		initFunction(admin)
	}

	admin.accessController.AddBeforeAction(func(request Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")

		request.SetData("admin_header_prefix", admin.prefix)
		request.SetData("javascripts", admin.javascripts)
		request.SetData("css", admin.css)
	})

	admin.accessController.AddAroundAction(
		createSessionAroundAction(
			admin.AppName,
			admin.Config.GetString("random"),
		),
	)

	googleAPIKey := admin.Config.GetStringWithFallback("google", "")
	admin.AdminController.AddBeforeAction(func(request Request) {
		request.SetData("google", googleAPIKey)
	})

	bindAPI(admin)
	admin.bindMigrationCommand()
	admin.initTemplates()

	must(admin.LoadTemplateFromFS(templatesFS, "templates/*.tmpl"))
	bindSystemstats(admin)
	admin.initRootActions()
	admin.initAutoRelations()
	bindDBBackupCron(admin)

	admin.AdminController.AddAroundAction(func(request Request, next func()) {
		session := request.GetData("session").(*sessions.Session)
		userID, ok := session.Values["user_id"].(int64)

		if !ok {
			request.Redirect(admin.GetURL("user/login"))
			return
		}

		var user User
		err := admin.Query().WhereIs("id", userID).Get(&user)
		if err != nil {
			request.Redirect(admin.GetURL("user/login"))
			return
		}

		randomness := admin.Config.GetString("random")
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", user.Locale)
		request.SetData("gravatar", user.gravatarURL())

		if !user.IsAdmin && !user.emailConfirmed() {
			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_confirmed"))
		}

		if !user.IsAdmin {
			var sysadmin User
			err := admin.Query().WhereIs("IsSysadmin", true).Get(&sysadmin)
			var sysadminEmail string
			if err == nil {
				sysadminEmail = sysadmin.Email
			}

			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_approved", sysadminEmail))
		}

		headerData := admin.getHeaderData(request)
		request.SetData("admin_header", headerData)
		request.SetData("main_menu", admin.getMainMenu(request))
		//request.SetData("admin_default_breadcrumbs", admin.createBreadcrumbs(user.Locale))

		next()
	})

	admin.AdminController.Get(admin.GetURL(""), func(request Request) {
		renderNavigationPage(request, adminNavigationPage{
			//Navigation:   admin.getAdminNavigation(GetUser(request), ""),
			PageTemplate: "admin_home_navigation",
			PageData:     admin.getHomeData(request),
		})
	})

	admin.AdminController.Get(admin.GetURL("_help/markdown"), func(request Request) {
		request.SetData("admin_yield", "admin_help_markdown")
		request.RenderView("admin_layout")
	})

	admin.AdminController.Get(admin.GetURL("_static/admin.js"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticAdminJS))
	})
	admin.AdminController.Get(admin.GetURL("_static/pikaday.js"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticPikadayJS))
	})
	admin.MainController().Get(admin.GetURL("_static/admin.css"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/css; charset=utf-8")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticAdminCSS))
	})

	admin.taskManager.init()

	for _, resource := range admin.resources {
		admin.initResource(resource)
	}

	bindSearch(admin)

	admin.AdminController.Get(admin.GetURL("*"), func(request Request) {
		render404(request)
	})

	admin.AddRole("sysadmin", admin.getSysadminPermissions())

	admin.resourcesInited = true
	return admin
}

//GetURL gets url
func (app App) GetURL(suffix string) string {
	ret := app.prefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//AddAction adds action
func (admin *App) AddAction(action Action) {
	admin.rootActions = append(admin.rootActions, action)
}

//AddFieldType adds field type
func (admin *App) AddFieldType(name string, fieldType FieldType) {
	_, exist := admin.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	admin.fieldTypes[name] = fieldType
}

//AddJavascript adds javascript
func (admin *App) AddJavascript(url string) {
	admin.javascripts = append(admin.javascripts, url)
}

//AddCSS adds CSS
func (admin *App) AddCSS(url string) {
	admin.css = append(admin.css, url)
}

//AddFlashMessage adds flash message to request
func AddFlashMessage(request Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	must(session.Save(request.Request(), request.Response()))
}

func addCurrentFlashMessage(request Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

func (admin *App) getResourceByName(name string) *Resource {
	return admin.resourceNameMap[columnName(name)]
}

func (admin *App) getDB() *sql.DB {
	return admin.db
}

//GetDB gets DB
func (admin *App) GetDB() *sql.DB {
	return admin.getDB()
}

func (admin *App) initRootActions() {
	for _, v := range admin.rootActions {
		bindAction(admin, nil, v, false)
	}
}

func (admin *App) initAutoRelations() {
	for _, v := range admin.resources {
		v.initAutoRelations()
	}
}

func (app *App) initTemplates() {
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
}

//GetItemURL gets item url
func (resource Resource) GetItemURL(item interface{}, suffix string) string {
	ret := resource.GetURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func render403(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}

func bindDBBackupCron(app *App) {
	app.NewTask("backup_db").SetHandler(
		func(tr *TaskActivity) error {
			err := BackupApp(app)
			if err != nil {
				return fmt.Errorf("Error while creating backup: %s", err)
			}
			return nil
		}).RepeatEvery(24 * time.Hour)

	app.NewTask("remove_old_backups").SetHandler(
		func(tr *TaskActivity) error {
			tr.SetStatus(0, fmt.Sprintf("Removing old backups"))
			deadline := time.Now().AddDate(0, 0, -7)
			backupPath := app.DotPath() + "/backups"
			files, err := ioutil.ReadDir(backupPath)
			if err != nil {
				return fmt.Errorf("Error while removing old backups: %s", err)
			}
			for _, file := range files {
				if file.ModTime().Before(deadline) {
					removePath := backupPath + "/" + file.Name()
					err := os.RemoveAll(removePath)
					if err != nil {
						return fmt.Errorf("Error while removing old backup file: %s", err)
					}
				}
			}
			app.Log().Println("Old backups removed")
			return nil
		}).RepeatEvery(1 * time.Hour)

}

func columnName(fieldName string) string {
	return utils.PrettyURL(fieldName)
}

//Unlocalized creates non localized name
func Unlocalized(name string) func(string) string {
	return func(string) string {
		return name
	}
}
