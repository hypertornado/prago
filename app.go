//Package prago is MVC framework for go
package prago

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/cachelib"
	"github.com/hypertornado/prago/messages"
	setup "github.com/hypertornado/prago/prago-setup/lib"
	"github.com/hypertornado/prago/utils"
)

//App is main struct of prago application
type App struct {
	codeName           string
	version            string
	DevelopmentMode    bool
	Config             config
	staticFilesHandler staticFilesHandler
	commands           *commands
	logger             *log.Logger
	templates          *templates
	mainController     *Controller
	Cache              *cachelib.Cache

	//App              *App
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

func NewTestingApp() *App {
	return createApp("__prago_test_app", "0.0", nil)
}

func createApp(codeName string, version string, initFunction func(*App)) *App {
	if codeName != "__prago_test_app" && !configExists(codeName) {
		if utils.ConsoleQuestion("File config.json does not exist. Can't start app. Would you like to start setup?") {
			setup.StartSetup(codeName)
		}
	}

	app := &App{
		codeName: codeName,
		version:  version,

		Config:   loadConfig(codeName),
		commands: &commands{},

		logger:         log.New(os.Stdout, "", log.LstdFlags),
		mainController: newMainController(),
		Cache:          cachelib.NewCache(),
	}

	app.initStaticFilesHandler()

	app.HumanName = app.codeName
	app.prefix = "/admin"
	app.resourceMap = make(map[reflect.Type]*Resource)
	app.resourceNameMap = make(map[string]*Resource)
	app.accessController = app.MainController().SubController()
	app.accessController.priorityRouter = true

	app.sendgridKey = app.Config.GetStringWithFallback("sendgridApi", "")
	app.noReplyEmail = app.Config.GetStringWithFallback("noReplyEmail", "")
	app.fieldTypes = make(map[string]FieldType)
	app.roles = make(map[string]map[string]bool)

	db, err := connectMysql(
		app.Config.GetStringWithFallback("dbUser", ""),
		app.Config.GetStringWithFallback("dbPassword", ""),
		app.Config.GetStringWithFallback("dbName", ""),
	)
	if err != nil {
		panic(err)
	}
	app.db = db

	app.AdminController = app.accessController.SubController()
	app.initDefaultFieldTypes()

	app.initTaskManager()

	app.CreateResource(User{}, initUserResource)
	app.CreateResource(Notification{}, initNotificationResource)
	app.CreateResource(File{}, initFilesResource)
	app.CreateResource(ActivityLog{}, initActivityLog)

	app.accessController.AddBeforeAction(func(request Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")

		request.SetData("locale", getLocale(request))

		request.SetData("admin_header_prefix", app.prefix)
		request.SetData("javascripts", app.javascripts)
		request.SetData("css", app.css)
	})

	app.accessController.AddAroundAction(
		app.createSessionAroundAction(
			app.Config.GetString("random"),
		),
	)

	googleAPIKey := app.Config.GetStringWithFallback("google", "")
	app.AdminController.AddBeforeAction(func(request Request) {
		request.SetData("google", googleAPIKey)
	})

	app.initAPI()
	app.initMigrationCommand()
	app.initTemplates()

	app.initSystemStats()
	app.initBackupCRON()

	app.AdminController.AddAroundAction(func(request Request, next func()) {
		session := request.GetData("session").(*sessions.Session)
		userID, ok := session.Values["user_id"].(int64)

		if !ok {
			request.Redirect(app.GetURL("user/login"))
			return
		}

		var user User
		err := app.Query().WhereIs("id", userID).Get(&user)
		if err != nil {
			request.Redirect(app.GetURL("user/login"))
			return
		}

		randomness := app.Config.GetString("random")
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", user.Locale)
		request.SetData("gravatar", user.gravatarURL())

		if !user.IsAdmin && !user.emailConfirmed() {
			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_confirmed"))
		}

		if !user.IsAdmin {
			var sysadmin User
			err := app.Query().WhereIs("IsSysadmin", true).Get(&sysadmin)
			var sysadminEmail string
			if err == nil {
				sysadminEmail = sysadmin.Email
			}

			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_approved", sysadminEmail))
		}

		headerData := app.getHeaderData(request)
		request.SetData("admin_header", headerData)
		request.SetData("main_menu", app.getMainMenu(request))

		next()
	})

	app.AdminController.Get(app.GetURL(""), func(request Request) {
		renderNavigationPage(request, adminNavigationPage{
			PageTemplate: "admin_home_navigation",
			PageData:     app.getHomeData(request),
		})
	})

	app.AdminController.Get(app.GetURL("_help/markdown"), func(request Request) {
		request.SetData("admin_yield", "admin_help_markdown")
		request.RenderView("admin_layout")
	})

	app.AdminController.Get(app.GetURL("_static/admin.js"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticAdminJS))
	})
	app.AdminController.Get(app.GetURL("_static/pikaday.js"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticPikadayJS))
	})
	app.MainController().Get(app.GetURL("_static/admin.css"), func(request Request) {
		request.Response().Header().Set("Content-type", "text/css; charset=utf-8")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(staticAdminCSS))
	})

	app.initSearch()

	if initFunction != nil {
		initFunction(app)
	}

	app.AdminController.Get(app.GetURL("*"), render404)

	app.initSysadminPermissions()
	app.initAllAutoRelations()

	return app
}

//NewApp creates App structure for prago app
func NewApp(appName, version string, initFunction func(*App)) {
	app := createApp(appName, version, initFunction)
	app.parseCommands()
}

func (app *App) initStaticFilesHandler() {
	paths := []string{}
	configValue, err := app.Config.Get("staticPaths")
	if err == nil {
		for _, path := range configValue.([]interface{}) {
			paths = append(paths, path.(string))
		}
	}
	app.staticFilesHandler = newStaticHandler(paths)
}

//Log returns logger structure
func (app App) Log() *log.Logger { return app.logger }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) dotPath() string { return os.Getenv("HOME") + "/." + app.codeName }

//ListenAndServe starts server on port
func (app *App) ListenAndServe(port int) error {
	app.Log().Printf("Server started: port=%d, pid=%d, developmentMode=%v\n", port, os.Getpid(), app.DevelopmentMode)

	if !app.DevelopmentMode {
		file, err := os.OpenFile(app.dotPath()+"/prago.log",
			os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
		must(err)
		app.logger.SetOutput(file)
	}

	return (&http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(port),
		Handler:        server{*app},
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}).ListenAndServe()
}

type server struct {
	app App
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.app.serveHTTP(w, r)
}

func (app App) serveHTTP(w http.ResponseWriter, r *http.Request) {
	request := Request{
		uuid:       utils.RandomString(10),
		receivedAt: time.Now(),
		w:          w,
		r:          r,
		app:        app,
		data:       nil,
	}
	w.Header().Set("X-Prago-Request", request.uuid)

	defer func() {
		if recoveryData := recover(); recoveryData != nil {
			app.recoveryFunction(request, recoveryData)
		}
	}()

	defer func() {
		app.writeAfterLog(request)
	}()

	if request.removeTrailingSlash() {
		return
	}

	if app.staticFilesHandler.serveStatic(request.Response(), request.Request()) {
		return
	}

	if !app.mainController.dispatchRequest(request) {
		request.Response().WriteHeader(http.StatusNotFound)
		request.Response().Write([]byte("404 — page not found (prago framework)"))
	}
}

func columnName(fieldName string) string {
	return utils.PrettyURL(fieldName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
