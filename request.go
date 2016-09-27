package prago

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
)

type Request struct {
	w    http.ResponseWriter
	r    *http.Request
	data map[string]interface{}
	app  *App
}

func (p *Request) Log() *logrus.Logger {
	return p.App().Log()
}

func (p *Request) IsProcessed() bool {
	_, processed := p.data["processed"]
	return processed
}

func (p *Request) SetProcessed()                     { p.data["processed"] = true }
func (p *Request) Request() (r *http.Request)        { return p.r }
func (p *Request) Response() (w http.ResponseWriter) { return p.w }
func (p *Request) Params() url.Values {
	return p.Request().Form
}
func (p *Request) SetData(k string, v interface{})        { p.data[k] = v }
func (p *Request) GetData(k string) interface{}           { return p.data[k] }
func (p *Request) AllRequestData() map[string]interface{} { return p.data }
func (p *Request) App() *App                              { return p.app }
func (p *Request) Header() http.Header                    { return p.w.Header() }

func newRequest(w http.ResponseWriter, r *http.Request, app *App) *Request {
	return &Request{
		w:    w,
		r:    r,
		app:  app,
		data: make(map[string]interface{}),
	}
}
