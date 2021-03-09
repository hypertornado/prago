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
)

//App is main struct of prago application
type App struct {
	codeName        string
	version         string
	development     *development
	DevelopmentMode bool
	Config          config
	staticFiles     staticFiles
	commands        *commands
	logger          *log.Logger
	templates       *templates
	mainController  *Controller
	Cache           *cachelib.Cache

	Logo             string
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

func NewTestingApp(initFunc func(*App)) *App {
	return createApp("__prago_test_app", "0.0", initFunc)
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
	app.resourceMap = make(map[reflect.Type]*Resource)
	app.resourceNameMap = make(map[string]*Resource)
	app.accessController = app.MainController().SubController()
	app.accessController.priorityRouter = true

	app.sendgridKey = app.Config.GetStringWithFallback("sendgridApi", "")
	app.noReplyEmail = app.Config.GetStringWithFallback("noReplyEmail", "")
	app.fieldTypes = make(map[string]FieldType)
	app.roles = make(map[string]map[string]bool)

	app.db = mustConnectDatabase(
		app.Config.GetStringWithFallback("dbUser", ""),
		app.Config.GetStringWithFallback("dbPassword", ""),
		app.Config.GetStringWithFallback("dbName", ""),
	)

	app.AdminController = app.accessController.SubController()
	app.initDefaultFieldTypes()
	app.initTaskManager()

	app.CreateResource(User{}, initUserResource)
	app.CreateResource(Notification{}, initNotificationResource)
	app.CreateResource(File{}, initFilesResource)
	app.CreateResource(ActivityLog{}, initActivityLog)

	app.initAPI()
	app.initDevelopment()
	app.initMigrationCommand()
	app.initTemplates()
	app.initSystemStats()
	app.initBackupCRON()
	app.initSearch()
	app.initAdminActions()

	if initFunction != nil {
		initFunction(app)
	}

	app.initAdminNotFoundAction()
	app.initSysadminPermissions()
	app.initAllAutoRelations()

	return app
}

//NewApp creates App structure for prago app
func NewApp(appName, version string, initFunction func(*App)) {
	app := createApp(appName, version, initFunction)
	app.parseCommands()
}

//Log returns logger structure
func (app App) Log() *log.Logger { return app.logger }

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
