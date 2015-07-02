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
	HttpIO() (w http.ResponseWriter, r *http.Request)
	Params() url.Values
	App() AppInterface
	SetData(string, interface{})
	GetData(string) interface{}
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

func (p *request) SetProcessed()                                    { p.data["processed"] = true }
func (p *request) HttpIO() (w http.ResponseWriter, r *http.Request) { return p.w, p.r }
func (p *request) Params() url.Values {
	_, r := p.HttpIO()
	return r.Form
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
	_, r := p.HttpIO()
	p.App().Log().Println(r.Method, r.URL.String())
}
