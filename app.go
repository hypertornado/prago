// Package prago is MVC framework for go
package prago

import (
	"database/sql"
	"os"
	"reflect"

	"github.com/hypertornado/prago/pragelastic"
)

// App is main struct of prago application
type App struct {
	codeName        string
	version         string
	development     *development
	developmentMode bool
	staticFiles     staticFiles
	commands        *commands
	logger          *logger
	templates       *templates
	cache           *cache
	sessionsManager *sessionsManager

	settings *settingsSingleton

	logo            []byte
	name            func(string) string
	resources       []resourceIface
	resourceMap     map[reflect.Type]resourceIface
	resourceNameMap map[string]resourceIface

	mainController   *controller
	appController    *controller
	accessController *controller
	adminController  *controller

	UsersResource       *Resource[user]
	FilesResource       *Resource[File]
	activityLogResource *Resource[activityLog]

	rootActions       []*Action
	db                *sql.DB
	sysadminTaskGroup *TaskGroup

	newsletters        *Newsletters
	notificationCenter *notificationCenter

	search        *adminSearch
	ElasticClient *pragelastic.Client

	fieldTypes    map[string]*fieldType
	javascripts   []string
	accessManager *accessManager

	apis []*API

	activityListeners []func(Activity)
	taskManager       *taskManager

	dashboardGroups []*DashboardGroup

	dbConfig *dbConnectConfig
}

func newTestingApp() *App {
	return createApp("__prago_test_app", "0.0")
}

func createApp(codeName string, version string) *App {
	app := &App{
		codeName:       codeName,
		version:        version,
		name:           unlocalized(codeName),
		commands:       &commands{},
		mainController: newMainController(),
		cache:          newCache(),
	}

	app.logger = newLogger(app)

	app.appController = app.mainController.subController()
	app.accessController = app.mainController.subController()
	app.accessController.priorityRouter = true
	app.adminController = app.accessController.subController()
	app.resourceMap = make(map[reflect.Type]resourceIface)
	app.resourceNameMap = make(map[string]resourceIface)
	app.fieldTypes = make(map[string]*fieldType)

	app.preInitTaskManager()
	app.initAccessManager()
	app.initDefaultFieldTypes()

	app.connectDB()

	app.initSettings()

	app.initElasticsearchClient()
	app.initLogger()

	app.initStaticFilesHandler()
	app.initNotifications()

	app.initUserResource()
	app.initFilesResource()

	app.initActivityLog()
	app.initHome()
	app.postInitTaskManager()
	app.initAdminActions()
	app.initBuild()
	app.initAPI()
	app.initDevelopment()
	app.initMigrationCommand()
	app.initTemplates()
	app.initSearch()
	app.initSystemStats()
	app.initSQLConsole()
	app.initSQLBackup()
	app.initBackupCRON()

	return app
}

func (app *App) Run() {
	app.afterInit()
	app.parseCommands()
}

func (app *App) afterInit() {
	app.initSessions()
	app.initDefaultResourceActions()
	app.bindAPIs()
	app.bindAllActions()
	app.initAdminNotFoundAction()
	app.initRelations()
}

func (app *App) initDefaultResourceActions() {
	for _, v := range app.resources {
		v.initDefaultResourceActions()
		v.initDefaultResourceAPIs()
	}
}

// New creates App structure for prago app
func New(appName, version string) *App {
	return createApp(appName, version)
}

func (app *App) GetDB() *sql.DB { return app.db }

// Log returns logger structure
func (app *App) Log() *logger { return app.logger }

// DevelopmentMode returns if app is running in development mode
func (app *App) DevelopmentMode() bool { return app.developmentMode }

// Name sets localized human name to app
func (app *App) Name(name func(string) string) *App {
	app.name = name
	return app
}

// Logo sets application public path to logo
func (app *App) Logo(logo []byte) *App {
	app.logo = logo
	return app
}

// DotPath returns path to hidden directory with app configuration and data
func (app *App) dotPath() string { return os.Getenv("HOME") + "/." + app.codeName }

func columnName(fieldName string) string {
	return prettyURL(fieldName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
