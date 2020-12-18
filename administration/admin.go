package administration

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/build"
	"github.com/hypertornado/prago/utils"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

//Administration is struct representing admin extension
type Administration struct {
	App              *prago.App
	Logo             string
	prefix           string
	HumanName        string
	resources        []*Resource
	resourceMap      map[reflect.Type]*Resource
	resourceNameMap  map[string]*Resource
	accessController *prago.Controller
	AdminController  *prago.Controller
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
func NewAdministration(app *prago.App, initFunction func(*Administration)) *Administration {
	admin := &Administration{
		App:              app,
		HumanName:        app.AppName,
		prefix:           "/admin",
		resourceMap:      make(map[reflect.Type]*Resource),
		resourceNameMap:  make(map[string]*Resource),
		accessController: app.MainController().SubController(),

		sendgridKey:  app.Config.GetStringWithFallback("sendgridApi", ""),
		noReplyEmail: app.Config.GetStringWithFallback("noReplyEmail", ""),

		fieldTypes: make(map[string]FieldType),
		roles:      make(map[string]map[string]bool),
	}

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
	admin.CreateResource(File{}, initFilesResource)
	admin.CreateResource(ActivityLog{}, initActivityLog)

	if initFunction != nil {
		initFunction(admin)
	}

	admin.accessController.AddBeforeAction(func(request prago.Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")

		request.SetData("admin_header_prefix", admin.prefix)
		request.SetData("javascripts", admin.javascripts)
		request.SetData("css", admin.css)
	})

	admin.accessController.AddAroundAction(
		createSessionAroundAction(
			admin.App.AppName,
			admin.App.Config.GetString("random"),
		),
	)

	googleAPIKey := admin.App.Config.GetStringWithFallback("google", "")
	admin.AdminController.AddBeforeAction(func(request prago.Request) {
		request.SetData("google", googleAPIKey)
	})

	bindAPI(admin)
	admin.bindMigrationCommand(admin.App)
	admin.initTemplates(admin.App)
	must(admin.App.LoadTemplateFromString(adminTemplates))
	bindSystemstats(admin)
	admin.initRootActions()
	admin.initAutoRelations()
	bindDBBackupCron(admin)

	admin.AdminController.AddAroundAction(func(request prago.Request, next func()) {
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

		randomness := admin.App.Config.GetString("random")
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

	admin.AdminController.Get(admin.GetURL(""), func(request prago.Request) {
		renderNavigationPage(request, adminNavigationPage{
			//Navigation:   admin.getAdminNavigation(GetUser(request), ""),
			PageTemplate: "admin_home_navigation",
			PageData:     admin.getHomeData(request),
		})
	})

	admin.AdminController.Get(admin.GetURL("_help/markdown"), func(request prago.Request) {
		request.SetData("admin_yield", "admin_help_markdown")
		request.RenderView("admin_layout")
	})

	admin.AdminController.Get(admin.GetURL("_static/admin.js"), func(request prago.Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(adminJS))
	})
	admin.AdminController.Get(admin.GetURL("_static/pikaday.js"), func(request prago.Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(pikadayJS))
	})
	admin.App.MainController().Get(admin.GetURL("_static/admin.css"), func(request prago.Request) {
		request.Response().Header().Set("Content-type", "text/css; charset=utf-8")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(adminCSS))
	})

	admin.taskManager.init()

	for _, resource := range admin.resources {
		admin.initResource(resource)
	}

	bindSearch(admin)

	admin.AdminController.Get(admin.GetURL("*"), func(request prago.Request) {
		render404(request)
	})

	admin.AddRole("sysadmin", admin.getSysadminPermissions())

	admin.resourcesInited = true
	return admin
}

//GetURL gets url
func (admin Administration) GetURL(suffix string) string {
	ret := admin.prefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//AddAction adds action
func (admin *Administration) AddAction(action Action) {
	admin.rootActions = append(admin.rootActions, action)
}

//AddFieldType adds field type
func (admin *Administration) AddFieldType(name string, fieldType FieldType) {
	_, exist := admin.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	admin.fieldTypes[name] = fieldType
}

//AddJavascript adds javascript
func (admin *Administration) AddJavascript(url string) {
	admin.javascripts = append(admin.javascripts, url)
}

//AddCSS adds CSS
func (admin *Administration) AddCSS(url string) {
	admin.css = append(admin.css, url)
}

//AddFlashMessage adds flash message to request
func AddFlashMessage(request prago.Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	must(session.Save(request.Request(), request.Response()))
}

func addCurrentFlashMessage(request prago.Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

func (admin *Administration) getResourceByName(name string) *Resource {
	return admin.resourceNameMap[columnName(name)]
}

func (admin *Administration) getDB() *sql.DB {
	return admin.db
}

//GetDB gets DB
func (admin *Administration) GetDB() *sql.DB {
	return admin.getDB()
}

func (admin *Administration) initRootActions() {
	for _, v := range admin.rootActions {
		bindAction(admin, nil, v, false)
	}
}

func (admin *Administration) initAutoRelations() {
	for _, v := range admin.resources {
		v.initAutoRelations()
	}
}

func (admin *Administration) initTemplates(app *prago.App) {
	app.AddTemplateFunction("markdown", func(text string) template.HTML {
		return template.HTML(markdown.New(markdown.Breaks(true)).RenderToString([]byte(text)))
	})

	app.AddTemplateFunction("message", func(language, id string) template.HTML {
		return template.HTML(messages.Messages.Get(language, id))
	})

	app.AddTemplateFunction("thumb", func(ids string) string {
		return admin.thumb(ids)
	})

	app.AddTemplateFunction("img", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			var image File
			err := admin.Query().WhereIs("uid", v).Get(&image)
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

func render403(request prago.Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request prago.Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}

func bindDBBackupCron(admin *Administration) {
	app := admin.App
	admin.NewTask("backup_db").SetHandler(
		func(tr *TaskActivity) error {
			err := build.BackupApp(app)
			if err != nil {
				return fmt.Errorf("Error while creating backup: %s", err)
			}
			return nil
		}).RepeatEvery(24 * time.Hour)

	admin.NewTask("remove_old_backups").SetHandler(
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
