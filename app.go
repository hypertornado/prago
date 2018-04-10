//Package prago is MVC framework for go
package prago

import (
	"github.com/Sirupsen/logrus"
	"github.com/hypertornado/prago/utils"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"strconv"
	"time"
)

//App is main struct of prago application
type App struct {
	AppName         string
	Version         string
	DevelopmentMode bool
	Config          config
	staticHandler   staticFilesHandler
	kingpin         *kingpin.Application
	commands        map[*kingpin.CmdClause]func(app *App)
	logger          *logrus.Logger
	cron            *cron
	templates       *templates
	mainController  *Controller
}

//NewApp creates App structure for prago app
func NewApp(appName, version string, initFunction func(*App)) {
	app := &App{
		AppName: appName,
		Version: version,

		Config: loadConfig(appName),

		cron:           newCron(),
		templates:      newTemplates(),
		mainController: newMainController(),
	}

	app.logger = createLogger(app.DotPath(), true)
	app.staticHandler = app.loadStaticHandler()

	initKingpinCommand(app)
	initFunction(app)

	commandName, err := app.kingpin.Parse(os.Args[1:])
	if err != nil {
		app.Log().Fatalf("cannot parse command name: %s", err)
	}

	for command, fn := range app.commands {
		if command.FullCommand() == commandName {
			fn(app)
			return
		}
	}
	app.Log().Fatalf("command not found: " + commandName)
}

func (app *App) loadStaticHandler() staticFilesHandler {
	paths := []string{}
	configValue, err := app.Config.Get("staticPaths")
	if err == nil {
		for _, path := range configValue.([]interface{}) {
			paths = append(paths, path.(string))
		}
	}
	return newStaticHandler(paths)
}

//Log returns logger structure
func (app *App) Log() *logrus.Logger { return app.logger }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) DotPath() string { return os.Getenv("HOME") + "/." + app.AppName }

//CreateCommand creates command for command line
func (app *App) CreateCommand(name, description string) *kingpin.CmdClause {
	return app.kingpin.Command(name, description)
}

//AddCommand adds function for command line command
func (app *App) AddCommand(cmd *kingpin.CmdClause, fn func(a *App)) {
	app.commands[cmd] = fn
}

//ListenAndServe starts server on port
func (app *App) ListenAndServe(port int, developmentMode bool) error {
	app.DevelopmentMode = developmentMode
	if !developmentMode {
		app.logger = createLogger(app.DotPath(), developmentMode)
	}

	server := &http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(port),
		Handler:        app,
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	app.Log().WithField("port", port).
		WithField("pid", os.Getpid()).
		WithField("development mode", app.DevelopmentMode).
		Info("Server started")

	return server.ListenAndServe()
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			recoveryFunction(request, recoveryData)
		}
	}()

	defer func() {
		request.writeAfterLog()
	}()

	if request.removeTrailingSlash() {
		return
	}

	if app.staticHandler.serveStatic(request.Response(), request.Request()) {
		return
	}

	if !app.mainController.dispatchRequest(request) {
		request.Response().WriteHeader(http.StatusNotFound)
		request.Response().Write([]byte("404 — not found"))
	}

}
