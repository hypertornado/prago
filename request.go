package prago

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
)

type Request interface {
	IsProcessed() bool
	SetProcessed()
	Params() url.Values
	App() *App
	SetData(string, interface{})
	GetData(string) interface{}
	Request() *http.Request
	Response() http.ResponseWriter
	AllRequestData() map[string]interface{}
	Header() http.Header
	Log() *logrus.Logger
}

type request struct {
	w    http.ResponseWriter
	r    *http.Request
	data map[string]interface{}
	app  *App
}

func (p *request) Log() *logrus.Logger {
	return p.App().data["logger"].(*logrus.Logger)
}

func (p *request) IsProcessed() bool {
	_, processed := p.data["processed"]
	return processed
}

func (p *request) SetProcessed()                     { p.data["processed"] = true }
func (p *request) Request() (r *http.Request)        { return p.r }
func (p *request) Response() (w http.ResponseWriter) { return p.w }
func (p *request) Params() url.Values {
	return p.Request().Form
}
func (p *request) SetData(k string, v interface{})        { p.data[k] = v }
func (p *request) GetData(k string) interface{}           { return p.data[k] }
func (p *request) AllRequestData() map[string]interface{} { return p.data }
func (p *request) App() *App                              { return p.app }
func (p *request) Header() http.Header                    { return p.w.Header() }

func newRequest(w http.ResponseWriter, r *http.Request, app *App) *request {
	return &request{
		w:    w,
		r:    r,
		app:  app,
		data: make(map[string]interface{}),
	}
}

func MiddlewareLogBefore(p Request) {
	p.App().data["logger"].(*logrus.Logger).Println(p.Request().Method, p.Request().URL.String())
}
