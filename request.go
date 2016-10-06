package prago

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
)

//Request represents structure for http request
type Request struct {
	w    http.ResponseWriter
	r    *http.Request
	data map[string]interface{}
	app  *App
}

//Log returns logger
func (p *Request) Log() *logrus.Logger {
	return p.App().Log()
}

//IsProcessed returns if request have been processed
func (p *Request) IsProcessed() bool {
	_, processed := p.data["processed"]
	return processed
}

//SetProcessed sets request as processed
func (p *Request) SetProcessed() { p.data["processed"] = true }

//Request returns underlying http.Request
func (p *Request) Request() (r *http.Request) { return p.r }

//Response returns underlying http.ResponseWriter
func (p *Request) Response() (w http.ResponseWriter) { return p.w }

//Params returns url.Values of request
func (p *Request) Params() url.Values {
	return p.Request().Form
}

//SetData sets request data
func (p *Request) SetData(k string, v interface{}) { p.data[k] = v }

//GetData returns request data
func (p *Request) GetData(k string) interface{} { return p.data[k] }

//AllRequestData returns all
func (p *Request) AllRequestData() map[string]interface{} { return p.data }

//App returns related app
func (p *Request) App() *App { return p.app }

//Header returns request header
func (p *Request) Header() http.Header { return p.w.Header() }
