//Package prago is MVC framework for go
package prago

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hypertornado/prago/utils"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

//TODO: dont make it global
var loggerMiddleware = &middlewareLogger{}

//App is main struct of prago application
type App struct {
	DevelopmentMode    bool
	Port               int
	StartedAt          time.Time
	RecoveryFunc       func(*Request, interface{})
	Config             *config
	data               map[string]interface{}
	requestMiddlewares []requestMiddleware
	middlewares        []Middleware
	kingpin            *kingpin.Application
	commands           map[*kingpin.CmdClause]func(app *App) error
	logger             *logrus.Logger
	dotPath            string
	cron               *cron
	mainController     *Controller
}

type requestMiddleware func(Request, func())

//NewApp creates App structure for prago app
func NewApp(appName, version string) *App {
	app := &App{
		RecoveryFunc:       defaultRecovery,
		data:               make(map[string]interface{}),
		requestMiddlewares: []requestMiddleware{},
		middlewares:        []Middleware{},
		dotPath:            os.Getenv("HOME") + "/." + appName,
	}
	app.mainController = newMainController(app)

	app.data["appName"] = appName
	app.data["version"] = version
	app.cron = newCron()

	app.AddMiddleware(middlewareCmd{})
	app.AddMiddleware(middlewareConfig{})
	app.AddMiddleware(loggerMiddleware)
	app.AddMiddleware(middlewareRemoveTrailingSlash)
	app.AddMiddleware(middlewareStatic{})
	app.AddMiddleware(middlewareParseRequest)
	app.AddMiddleware(middlewareView{})
	app.AddMiddleware(middlewareDispatcher{})
	return app
}

//Log returns logger structure
func (app *App) Log() *logrus.Logger { return app.logger }

//DotPath returns path to hidden directory with app configuration and data
func (app *App) DotPath() string { return app.dotPath }

//CreateCommand creates command for command line
func (app *App) CreateCommand(name, description string) *kingpin.CmdClause {
	return app.kingpin.Command(name, description)
}

//AddCommand adds function for command line command
func (app *App) AddCommand(cmd *kingpin.CmdClause, fn func(a *App) error) {
	app.commands[cmd] = fn
}

//AddMiddleware adds optional middlewares
func (app *App) AddMiddleware(m Middleware) {
	app.middlewares = append(app.middlewares, m)
}

//Data returns map of all app data
func (app *App) Data() map[string]interface{} {
	return app.data
}

func (app *App) initMiddlewares() error {
	for _, middleware := range app.middlewares {
		if err := middleware.Init(app); err != nil {
			return fmt.Errorf("initializating middleware: %s", err)
		}
	}
	return nil
}

//Init runs all middleware init function
func (app *App) Init() error {
	err := app.initMiddlewares()
	if err != nil {
		return fmt.Errorf("initializating middlewares: %s", err)
	}

	commandName, err := app.kingpin.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parsing command name %s", err)
	}

	for command, fn := range app.commands {
		if command.FullCommand() == commandName {
			err := fn(app)
			if err != nil {
				return fmt.Errorf("running command name %s: %s", commandName, err)
			}
			return nil
		}
	}

	return errors.New("command not found: " + commandName)
}

func (app *App) route(m method, path string, controller *Controller, routeAction func(p Request), constraints ...Constraint) error {
	router := app.data["router"].(*router)
	if router == nil {
		return errors.New("couldnt find router")
	}

	route := newRoute(m, path, controller, routeAction, constraints)
	router.addRoute(route)
	return nil
}

//ListenAndServe starts server on port
func (app *App) ListenAndServe(port int, developmentMode bool) error {
	app.DevelopmentMode = developmentMode
	app.Port = port
	app.StartedAt = time.Now()

	if developmentMode {
		loggerMiddleware.setStdOut()
	}

	server := &http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(port),
		Handler:        app,
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	app.writeStartInfo()
	return server.ListenAndServe()
}

func (app *App) writeStartInfo() error {
	pid := os.Getpid()

	err := ioutil.WriteFile(
		app.dotPath+"/last.pid",
		[]byte(fmt.Sprintf("%d", pid)),
		0777,
	)
	if err != nil {
		return err
	}

	fmt.Printf("Server started\nport: %d\npid: %d\ndevelopment mode: %t\n", app.Port, pid, app.DevelopmentMode)

	app.Log().WithField("port", app.Port).
		WithField("pid", pid).
		WithField("development mode", app.DevelopmentMode).
		Info("Server started")

	return nil
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &Request{
		w:    w,
		r:    r,
		app:  app,
		data: make(map[string]interface{}),
	}

	defer func() {
		if recoveryData := recover(); recoveryData != nil {
			app.RecoveryFunc(request, recoveryData)
		}
	}()

	callRequestMiddlewares(request, app.requestMiddlewares)
}

func callRequestMiddlewares(request *Request, middlewares []requestMiddleware) {
	f := func() {}
	for i := len(middlewares) - 1; i >= 0; i-- {
		j := i
		prevF := f
		f = func() {
			middlewares[j](*request, prevF)
		}
	}
	f()
}

func defaultRecovery(p *Request, recoveryData interface{}) {
	uuid := utils.RandomString(10)

	p.Response().WriteHeader(500)
	p.Response().Write([]byte("500 Internal Server Error, errorid " + uuid))
	p.Log().Errorln(fmt.Sprintf("500 Internal Server Error, errorid %s\n%s\nstack:\n", uuid, recoveryData))
	p.Log().Errorln(string(debug.Stack()))
}
