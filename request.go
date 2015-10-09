package prago

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"net/http"
	"net/url"
)

type Request interface {
	IsProcessed() bool
	SetProcessed()
	Params() url.Values
	App() AppInterface
	SetData(string, interface{})
	GetData(string) interface{}
	Request() *http.Request
	Response() http.ResponseWriter
	AllRequestData() map[string]interface{}
	Header() http.Header
	Session() *sessions.Session
	Log() *logrus.Logger
}

type request struct {
	w    http.ResponseWriter
	r    *http.Request
	data map[string]interface{}
	app  AppInterface
}

func (p *request) Log() *logrus.Logger {
	return p.App().Log()
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
func (p *request) App() AppInterface                      { return p.app }
func (p *request) Header() http.Header                    { return p.w.Header() }
func (p *request) Session() *sessions.Session {
	session, ok := p.GetData("session").(*sessions.Session)
	if !ok {
		panic("can't get session")
	}
	return session
}

func newRequest(w http.ResponseWriter, r *http.Request, app AppInterface) *request {
	return &request{
		w:    w,
		r:    r,
		app:  app,
		data: make(map[string]interface{}),
	}
}

func MiddlewareLogBefore(p Request) {
	p.App().Log().Println(p.Request().Method, p.Request().URL.String())
}
