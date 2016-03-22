package prago

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hypertornado/prago/utils"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

type App struct {
	data               map[string]interface{}
	events             *Events
	requestMiddlewares []RequestMiddleware
	middlewares        []Middleware
}

type RequestMiddleware func(Request, func())

func NewApp(name string) *App {
	app := &App{
		data:               make(map[string]interface{}),
		events:             NewEvents(),
		requestMiddlewares: []RequestMiddleware{},
		middlewares:        []Middleware{},
	}

	app.data["logger"] = utils.DefaultLogger()
	app.data["mainController"] = newMainController(app)
	app.data["appName"] = name
	app.data["router"] = NewRouter()

	app.AddMiddleware(MiddlewareConfig{})
	app.AddMiddleware(MiddlewareLogBefore)
	app.AddMiddleware(MiddlewareRemoveTrailingSlash)
	app.AddMiddleware(MiddlewareStatic)
	app.AddMiddleware(MiddlewareParseRequest)
	app.AddMiddleware(MiddlewareView{})
	app.AddMiddleware(MiddlewareDispatcher)

	return app
}

func (a *App) AddMiddleware(m Middleware) {
	a.middlewares = append(a.middlewares, m)
}

func (a *App) Data() map[string]interface{} {
	return a.data
}

func (a *App) Init(init func(*App)) error {
	for _, middleware := range a.middlewares {
		if err := middleware.Init(a); err != nil {
			return err
		}
	}
	return a.cmd(init)
}

func (a *App) MainController() (ret *Controller) {
	ret = a.data["mainController"].(*Controller)
	if ret == nil {
		panic("couldnt find controller")
	}
	return
}

func (a *App) Route(m method, path string, controller *Controller, action func(p Request), constraints ...Constraint) error {
	router := a.data["router"].(*Router)
	if router == nil {
		return errors.New("couldnt find router")
	}

	bindedAction := controller.NewAction(action)
	route := NewRoute(m, path, bindedAction, constraints)
	router.AddRoute(route)
	return nil
}

func (a *App) ListenAndServe(port int, developmentMode bool) error {
	a.data["developmentMode"] = developmentMode

	server := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        a,
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	utils.WriteStartInfo(a.data["logger"].(*logrus.Logger), port, developmentMode)
	return server.ListenAndServe()
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
	p.Response().WriteHeader(500)

	developmentMode := p.App().data["developmentMode"].(bool)

	if developmentMode {
		p.Response().Write([]byte(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData)))
		p.Response().Write(debug.Stack())
	} else {
		p.Response().Write([]byte("We are sorry, some error occured. (500)"))
	}
	p.Log().Errorln(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData))
	p.Log().Errorln(string(debug.Stack()))
}

func Redirect(request Request, urlStr string) {
	request.Header().Set("Location", urlStr)
	request.Response().WriteHeader(http.StatusMovedPermanently)
}
