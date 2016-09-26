package prago

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

var loggerMiddleware = &MiddlewareLogger{}

type App struct {
	DevelopmentMode    bool
	Port               int
	StartedAt          time.Time
	data               map[string]interface{}
	requestMiddlewares []RequestMiddleware
	middlewares        []Middleware
	kingpin            *kingpin.Application
	commands           map[*kingpin.CmdClause]func(app *App) error
	logger             *logrus.Logger
	dotPath            string
	cron               *cron
}

type RequestMiddleware func(Request, func())

func NewApp(appName, version string) *App {
	app := &App{
		data:               make(map[string]interface{}),
		requestMiddlewares: []RequestMiddleware{},
		middlewares:        []Middleware{},
		dotPath:            os.Getenv("HOME") + "/." + appName,
	}

	app.data["mainController"] = newMainController(app)
	app.data["appName"] = appName
	app.data["version"] = version
	app.cron = newCron()

	app.AddMiddleware(MiddlewareCmd{})
	app.AddMiddleware(MiddlewareConfig{})
	app.AddMiddleware(loggerMiddleware)
	app.AddMiddleware(MiddlewareRemoveTrailingSlash)
	app.AddMiddleware(MiddlewareStatic{})
	app.AddMiddleware(MiddlewareParseRequest)
	app.AddMiddleware(MiddlewareView{})
	app.AddMiddleware(MiddlewareDispatcher{})

	return app
}

func (a *App) Log() *logrus.Logger { return a.logger }
func (a *App) DotPath() string     { return a.dotPath }

func (a *App) AddCommand(cmd *kingpin.CmdClause, fn func(app *App) error) {
	a.commands[cmd] = fn
}

func (a *App) CreateCommand(name, description string) *kingpin.CmdClause {
	return a.kingpin.Command(name, description)
}

func (a *App) AddMiddleware(m Middleware) {
	a.middlewares = append(a.middlewares, m)
}

func (a *App) Data() map[string]interface{} {
	return a.data
}

func (a *App) initMiddlewares() error {
	for _, middleware := range a.middlewares {
		if err := middleware.Init(a); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Init() error {
	err := a.initMiddlewares()
	if err != nil {
		return err
	}

	commandName, err := a.kingpin.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	for command, fn := range a.commands {
		if command.FullCommand() == commandName {
			return fn(a)
		}
	}

	return errors.New("command not found: " + commandName)
}

func (app *App) route(m method, path string, controller *Controller, action func(p Request), constraints ...Constraint) error {
	router := app.data["router"].(*router)
	if router == nil {
		return errors.New("couldnt find router")
	}

	bindedAction := controller.NewAction(action)
	route := newRoute(m, path, bindedAction, constraints)
	router.addRoute(route)
	return nil
}

func (app *App) ListenAndServe(port int, developmentMode bool) error {
	app.DevelopmentMode = developmentMode
	app.Port = port
	app.StartedAt = time.Now()

	if developmentMode {
		loggerMiddleware.setStdOut()
	}

	server := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
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

	developmentModeStr := "false"
	if app.DevelopmentMode {
		developmentModeStr = "true"
	}
	fmt.Printf("Server started\nport: %d\npid: %d\ndevelopment mode: %s\n", app.Port, pid, developmentModeStr)

	app.Log().WithField("port", app.Port).
		WithField("pid", pid).
		WithField("development mode", app.DevelopmentMode).
		Info("Server started")

	return nil
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, app)
}

func handleRequest(w http.ResponseWriter, r *http.Request, app *App) {
	request := newRequest(w, r, app)

	defer func() {
		if recoveryData := recover(); recoveryData != nil {
			recoveryFromServerError(request, recoveryData)
		}
	}()

	callRequestMiddlewares(request, app.requestMiddlewares)
}

func callRequestMiddlewares(request Request, middlewares []RequestMiddleware) {
	f := func() {}
	for i := len(middlewares) - 1; i >= 0; i-- {
		j := i
		prevF := f
		f = func() {
			middlewares[j](request, prevF)
		}
	}
	f()
}

func recoveryFromServerError(p Request, recoveryData interface{}) {
	if p.App().DevelopmentMode {
		p.Response().Header().Set("Content-Type", "text/plain")
		p.Response().WriteHeader(500)
		p.Response().Write([]byte(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData)))
		p.Response().Write(debug.Stack())
	} else {
		p.Response().WriteHeader(500)
		p.Response().Write([]byte("We are sorry, some error occured. (500)"))
	}
	p.Log().Errorln(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData))
	p.Log().Errorln(string(debug.Stack()))
}
