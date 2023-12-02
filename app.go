// Package prago is MVC framework for go
package prago

import (
	"database/sql"
	"embed"
	"os"
	"reflect"
	"sync"
)

//TODO: implement https://pkg.go.dev/expvar#Handler
//https://github.com/shirou/gopsutil

//https://github.com/divan/expvarmon

// App is main struct of prago application
type App struct {
	codeName        string
	version         string
	icon            string
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
	resources       []*resourceData
	resourceMap     map[reflect.Type]*resourceData
	resourceNameMap map[string]*resourceData

	mainController   *controller
	appController    *controller
	accessController *controller
	adminController  *controller

	UsersResource      *Resource[user]
	userDataCache      map[int64]*userData
	userDataCacheMutex *sync.RWMutex

	FilesResource       *Resource[File]
	activityLogResource *Resource[activityLog]

	rootActions []*Action
	db          *sql.DB
	//sysadminTaskGroup *TaskGroup

	newsletters        *Newsletters
	notificationCenter *notificationCenter

	//search        *adminSearch
	//ElasticClient *pragelastic.Client

	fieldTypes    map[string]*fieldType
	javascripts   []string
	accessManager *accessManager

	apis []*API

	activityListeners []func(Activity)
	taskManager       *taskManager

	dashboardTableMap  map[string]*DashboardTable
	dashboardFigureMap map[string]*DashboardFigure

	MainBoard *Board

	dbConfig *dbConnectConfig

	iconsFS     *embed.FS
	iconsPrefix string
}

const testingAppName = "__prago_test_app"

func newTestingApp() *App {
	return createApp(testingAppName, "0.0")
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
	app.resourceMap = make(map[reflect.Type]*resourceData)
	app.resourceNameMap = make(map[string]*resourceData)
	app.fieldTypes = make(map[string]*fieldType)

	app.preInitTaskManager()
	app.initAccessManager()
	app.initDefaultFieldTypes()

	var testing bool
	if app.codeName == testingAppName {
		testing = true
	}
	app.connectDB(testing)

	app.initUserDataCache()
	app.initBoard()
	app.initSettings()
	//app.initElasticsearchClient()
	app.initLogger()
	app.initStaticFilesHandler()
	app.initNotifications()

	app.initUserResource()
	app.initFilesResource()

	app.initActivityLog()
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
	app.initDashboard()
	app.initIcons()
	app.initMenuAPI()

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
	for _, resourceData := range app.resources {
		resourceData.initDefaultResourceActions()
		resourceData.initDefaultResourceAPIs()
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

func (app *App) Icon(icon string) *App {
	app.icon = icon
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
