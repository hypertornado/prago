package prago

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type AppInterface interface {
	Log() *logrus.Logger
	Templates() *template.Template
	Router() *Router
	Route(method method, path string, controller *Controller, fn func(p Request), c ...Constraint)
	MainController() *Controller
	SessionStore() sessions.Store
	DevelopmentMode() bool
}

type App struct {
	sessionStore    sessions.Store
	log             *logrus.Logger
	developmentMode bool
	templates       *template.Template
	router          *Router
	mainController  *Controller
	middlewares     []Middleware
}

func NewApp() *App {
	return &App{
		log:            defaultLogger(),
		templates:      nil,
		router:         NewRouter(),
		mainController: newController(nil),
		middlewares: []Middleware{
			MiddlewareLogBefore,
			MiddlewareParseRequest,
			MiddlewareInitSession,
			MiddlewareStatic,
			MiddlewareDispatcher,
			MiddlewareSaveSession,
			MiddlewareWriteResponse,
		},
	}
}

func (h *App) DevelopmentMode() bool                     { return h.developmentMode }
func (h *App) AddSessionStore(store sessions.Store)      { h.sessionStore = store }
func (h *App) AddTemplates(templates *template.Template) { h.templates = templates }
func (h *App) Log() *logrus.Logger                       { return h.log }
func (h *App) Templates() *template.Template             { return h.templates }
func (h *App) Router() *Router                           { return h.router }
func (h *App) SessionStore() sessions.Store              { return h.sessionStore }
func (h *App) MainController() *Controller               { return h.mainController }

func (h *App) Route(m method, path string, controller *Controller, action func(p Request), constraints ...Constraint) {
	bindedAction := controller.NewAction(action)
	route := NewRoute(m, path, bindedAction, constraints)
	h.Router().AddRoute(route)
}

func (h *App) ListenAndServe(port int, developmentMode bool) error {

	h.developmentMode = developmentMode

	server := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        h,
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	writeStartInfo(h.Log(), port, developmentMode)
	return server.ListenAndServe()
}

type Middleware func(Request)

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

	for _, middleware := range app.middlewares {
		middleware(request)
	}
}

func recoveryFromServerError(p Request, recoveryData interface{}) {
	w, _ := p.HttpIO()
	w.WriteHeader(500)
	if p.App().DevelopmentMode() {
		w.Write([]byte(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData)))
		w.Write(debug.Stack())
	} else {
		w.Write([]byte("We are sorry, some error occured. Admin has been contacted. (500)"))
	}
}

func MiddlewareParseRequest(p Request) {
	_, r := p.HttpIO()

	contentType := r.Header.Get("Content-Type")

	var err error

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err = r.ParseMultipartForm(1000000)
		Must(err)

		for k, values := range r.MultipartForm.Value {
			for _, v := range values {
				r.Form.Add(k, v)
			}
		}
	} else {
		err = r.ParseForm()
		Must(err)
	}
}
