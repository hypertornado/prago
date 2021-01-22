//Package prago is MVC framework for go
package prago

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hypertornado/prago/cachelib"
	setup "github.com/hypertornado/prago/prago-setup/lib"
	"github.com/hypertornado/prago/utils"
)

//App is main struct of prago application
type App struct {
	AppName         string
	Version         string
	DevelopmentMode bool
	Config          config
	staticHandler   staticFilesHandler
	commands        *commands
	logger          *log.Logger
	templates       *templates
	mainController  *Controller
	Cache           *cachelib.Cache
}

func NewTestingApp() *App {
	return createApp("__prago_test_app", "0.0", nil)
}

func createApp(appName string, version string, initFunction func(*App)) *App {
	if appName != "__prago_test_app" && !configExists(appName) {
		if utils.ConsoleQuestion("File config.json does not exist. Can't start app. Would you like to start setup?") {
			setup.StartSetup(appName)
		}
	}

	app := &App{
		AppName: appName,
		Version: version,

		Config:   loadConfig(appName),
		commands: &commands{},

		logger:         log.New(os.Stdout, "", log.LstdFlags),
		templates:      newTemplates(),
		mainController: newMainController(),
		Cache:          cachelib.NewCache(),
	}

	app.staticHandler = app.loadStaticHandler()
	if initFunction != nil {
		initFunction(app)
	}
	return app
}

//NewApp creates App structure for prago app
func NewApp(appName, version string, initFunction func(*App)) {
	app := createApp(appName, version, initFunction)
	app.parseCommands()
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
func (app App) Log() *log.Logger { return app.logger }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) DotPath() string { return os.Getenv("HOME") + "/." + app.AppName }

//ListenAndServe starts server on port
func (app *App) ListenAndServe(port int) error {
	app.Log().Printf("Server started: port=%d, pid=%d, developmentMode=%v\n", port, os.Getpid(), app.DevelopmentMode)

	if !app.DevelopmentMode {
		file, err := os.OpenFile(app.DotPath()+"/prago.log",
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
		request.Response().Write([]byte("404 — page not found (prago framework)"))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
