// Package prago is MVC framework for go
package prago

import (
	"database/sql"
	"embed"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
)

//TODO: implement https://pkg.go.dev/expvar#Handler
//https://github.com/shirou/gopsutil

//https://github.com/divan/expvarmon

type App struct {
	testing         bool
	port            int
	codeName        string
	version         string
	icon            string
	development     *development
	developmentMode bool
	staticFiles     staticFiles
	commands        []*command
	logger          *logger
	adminTemplates  *PragoTemplates
	cache           *cache
	sessionsManager *sessionsManager

	settings *settingsSingleton

	logo            []byte
	name            func(string) string
	resources       []*Resource
	resourceMap     map[reflect.Type]*Resource
	resourceNameMap map[string]*Resource

	mainController   *controller
	appController    *controller
	accessController *controller
	adminController  *controller

	UsersResource      *Resource
	userDataCache      map[int64]*userData
	userDataCacheMutex *sync.RWMutex

	FilesResource       *Resource
	activityLogResource *Resource

	rootActions []*Action
	db          *sql.DB

	notificationCenter *notificationCenter

	fieldTypes    map[string]*fieldType
	accessManager *accessManager

	apis []*API

	activityListeners []func(Activity)

	dashboardTableMap    map[string]*dashboardTable
	dashboardFigureMap   map[string]*dashboardFigure
	dashboardTimelineMap map[string]*Timeline

	MainBoard *Board

	dbConfig *dbConnectConfig

	iconsFS     *embed.FS
	iconsPrefix string

	customSearchFunctions []func(string, UserData) []*CustomSearchResult

	tasksMap map[string]*Task

	cronTasks []*cronTask

	logHandler func(string, string)

	router *router

	EmailSentHandler func(*Email)

	serverSetup func(*http.Server)

	afterRequestServedHandler func(*Request)
}

func NewTesting(t *testing.T, initHandler func(app *App)) *App {
	app := createApp("__prago_test_app", "0.0", true)
	initHandler(app)
	app.afterInit()
	app.unsafeDropTables()
	app.migrate(false)
	return app
}

func createApp(codeName string, version string, testing bool) *App {
	app := &App{
		testing:  testing,
		codeName: codeName,
		version:  version,
		name:     unlocalized(codeName),
		cache:    newCache(),
		router:   newRouter(),
	}

	app.logger = newLogger(app)

	app.mainController = newMainController(app)
	app.appController = app.mainController.subController()
	app.accessController = app.mainController.subController()
	app.accessController.priorityRouter = true
	app.adminController = app.accessController.subController()

	app.resourceMap = make(map[reflect.Type]*Resource)
	app.resourceNameMap = make(map[string]*Resource)
	app.fieldTypes = make(map[string]*fieldType)

	app.initPprof()

	app.initAccessManager()
	app.initDefaultFieldTypes()

	app.connectDB(testing)

	app.initUserDataCache()
	app.initBoard()
	app.initSettings()
	app.initStaticFilesHandler()
	app.initNotifications()

	app.initUserResource()
	app.initFilesResource()
	app.initEmailSentResource()

	app.initSystemStats()
	app.initActivityLog()
	app.postInitTaskManager()
	app.initAdminActions()
	app.initBuild()
	app.initAPI()
	app.initDevelopment()
	app.initMigrationCommand()
	app.initTemplates()
	//app.initElasticsearch()
	app.initSearch()
	app.initSQLConsole()
	app.initSQLBackup()
	app.initBackupCRON()
	app.initDashboard()
	app.initIcons()
	app.initMenuAPI()
	app.initCron()
	app.initCacheStats()
	app.initMailing()

	return app
}

func (app *App) Run() {
	app.afterInit()
	app.parseCommands()
}

func (app *App) afterInit() {
	app.afterInitUserResource()
	app.afterInitFilesResource()

	for _, resource := range app.resources {
		resource.afterInit()
	}

	app.initSessions()
	app.initDefaultResourceActions()
	app.initAPIs()
	app.initAllActions()
	app.initAdminNotFoundAction()
	app.initRelations()
	app.initMultipleItemActions()
	app.initFieldValidations()
}

func (app *App) initDefaultResourceActions() {
	for _, resource := range app.resources {
		resource.initDefaultResourceActions()
		resource.initDefaultResourceAPIs()
	}
}

// New creates App structure for prago app
func New(appName, version string) *App {
	return createApp(appName, version, false)
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

func (app *App) BaseURL() string {
	return app.mustGetSetting("base_url")
}

func (app *App) GoogleKey() string {
	return app.mustGetSetting("google_key")
}

func (app *App) IconURL() string {
	ret := app.mustGetSetting("icon_image_url")
	if ret != "" {
		return ret
	}
	return app.BaseURL() + "/admin/logo"
}

func (app *App) RandomizationString() string {
	return app.mustGetSetting("random")
}

// DotPath returns path to hidden directory with app configuration and data
func (app *App) dotPath() string { return os.Getenv("HOME") + "/." + app.codeName }

func (app *App) SetAfterRequestServedHandler(fn func(*Request)) {
	app.afterRequestServedHandler = fn
}

func columnName(fieldName string) string {
	return prettyURL(fieldName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func (app *App) getVersionString() string {
	if app.DevelopmentMode() {
		return fmt.Sprintf("%s-development-%d", app.version, rand.Intn(10000000000))
	}
	return app.version
}
