//Package prago is MVC framework for go
package prago

import (
	"github.com/Sirupsen/logrus"
	"github.com/hypertornado/prago/utils"
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
	commands        []*command
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

	initFunction(app)
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
func (app App) Log() *logrus.Logger { return app.logger }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) DotPath() string { return os.Getenv("HOME") + "/." + app.AppName }

//ListenAndServe starts server on port
func (app *App) ListenAndServe(port int, developmentMode bool) error {
	app.Log().WithField("port", port).
		WithField("pid", os.Getpid()).
		WithField("development mode", app.DevelopmentMode).
		Info("Server started")

	app.DevelopmentMode = developmentMode
	if !developmentMode {
		app.logger = createLogger(app.DotPath(), developmentMode)
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

	request.Log().
		WithField("uuid", request.uuid).
		WithField("took", time.Now().Sub(request.receivedAt)).
		Println("before dispatch")

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
