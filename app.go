//Package prago is MVC framework for go
package prago

import (
	"database/sql"
	"log"
	"os"
	"reflect"

	"github.com/hypertornado/prago/cachelib"
	setup "github.com/hypertornado/prago/prago-setup/lib"
	"github.com/hypertornado/prago/utils"
	"github.com/sendgrid/sendgrid-go"
)

//App is main struct of prago application
type App struct {
	codeName        string
	version         string
	development     *development
	developmentMode bool
	config          config
	staticFiles     staticFiles
	commands        *commands
	logger          *log.Logger
	templates       *templates
	cache           *cachelib.Cache
	sessionsManager *sessionsManager

	logo            string
	name            func(string) string
	resources       []*Resource
	resourceMap     map[reflect.Type]*Resource
	resourceNameMap map[string]*Resource

	mainController   *controller
	appController    *controller
	accessController *controller
	adminController  *controller

	UsersResource *Resource
	FilesResource *Resource

	rootActions    []*Action
	db             *sql.DB
	sendgridClient *sendgrid.Client

	sendgridKey  string
	noReplyEmail string

	newsletter *newsletterMiddleware

	search *adminSearch

	fieldTypes  map[string]*FieldType
	javascripts []string
	css         []string
	//roles       map[string]map[string]bool
	accessManager *accessManager

	apis []*API

	activityListeners []func(ActivityLog)
	taskManager       *taskManager
}

func newTestingApp() *App {
	return createApp("__prago_test_app", "0.0")
}

func createApp(codeName string, version string) *App {
	if codeName != "__prago_test_app" && !configExists(codeName) {
		if utils.ConsoleQuestion("File config.json does not exist. Can't start app. Would you like to start setup?") {
			setup.StartSetup(codeName)
		}
	}

	app := &App{
		codeName: codeName,
		version:  version,

		commands: &commands{},

		logger:         log.New(os.Stdout, "", log.LstdFlags),
		mainController: newMainController(),
		cache:          cachelib.NewCache(),
	}

	app.appController = app.mainController.subController()
	app.accessController = app.mainController.subController()
	app.accessController.priorityRouter = true

	app.initConfig()
	app.initEmail()
	app.initSessions()
	app.initAccessManager()
	app.initStaticFilesHandler()

	app.name = Unlocalized(app.codeName)
	app.resourceMap = make(map[reflect.Type]*Resource)
	app.resourceNameMap = make(map[string]*Resource)

	app.fieldTypes = make(map[string]*FieldType)

	app.db = mustConnectDatabase(
		app.ConfigurationGetStringWithFallback("dbUser", ""),
		app.ConfigurationGetStringWithFallback("dbPassword", ""),
		app.ConfigurationGetStringWithFallback("dbName", ""),
	)

	app.adminController = app.accessController.subController()
	app.initDefaultFieldTypes()
	app.initUserResource()
	initNotificationResource(
		app.Resource(
			Notification{},
		),
	)
	app.initFilesResource()
	initActivityLog(
		app.Resource(
			ActivityLog{},
		),
	)

	app.initHome()
	app.initTaskManager()
	app.initAdminActions()
	app.initBuild()
	app.initAPI()
	app.initDevelopment()
	app.initMigrationCommand()
	app.initTemplates()
	app.initSearch()
	app.initSystemStats()
	app.initBackupCRON()
	return app
}

func (app *App) afterInit() {
	app.initDefaultResourceActions()
	app.bindAPIs()
	app.bindAllActions()
	app.initAdminNotFoundAction()
	app.initAllAutoRelations()
}

func (app *App) Run() {
	app.afterInit()
	app.parseCommands()
}

//Application creates App structure for prago app
func Application(appName, version string) *App {
	return createApp(appName, version)
}

//Log returns logger structure
func (app *App) Log() *log.Logger { return app.logger }

//DevelopmentMode returns if app is running in development mode
func (app *App) DevelopmentMode() bool { return app.developmentMode }

//SetName sets localized human name to app
func (app *App) SetName(name func(string) string) { app.name = name }

//SetLogoPath sets application public path to logo
func (app *App) SetLogoPath(logo string) { app.logo = logo }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) dotPath() string { return os.Getenv("HOME") + "/." + app.codeName }

func columnName(fieldName string) string {
	return utils.PrettyURL(fieldName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
