package prago

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//Request represents structure for http request
type Request struct {
	uuid       string
	receivedAt time.Time
	w          http.ResponseWriter
	r          *http.Request
	data       map[string]interface{}
	app        *App
}

//Log returns logger
func (p *Request) Log() *logrus.Logger {
	return p.App().Log()
}

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

//Render outputs view of viewName with statusCode to request
func Render(request Request, statusCode int, viewName string) {
	request.Header().Add("Content-Type", "text/html")
	request.Response().WriteHeader(statusCode)
	Must(
		request.app.templates.templates.ExecuteTemplate(
			request.Response(),
			viewName,
			request.AllRequestData(),
		),
	)
}

func (request Request) RenderJSON(data interface{}) {
	request.RenderJSONWithCode(data, 200)
}

func (request Request) RenderJSONWithCode(data interface{}, code int) {
	request.Response().Header().Add("Content-type", "application/json")

	pretty := false
	if request.Params().Get("pretty") == "true" {
		pretty = true
	}

	var responseToWrite interface{}
	if code >= 400 {
		responseToWrite = map[string]interface{}{"error": data, "errorCode": code}
	} else {
		responseToWrite = data
	}

	var result []byte
	var e error

	if pretty == true {
		result, e = json.MarshalIndent(responseToWrite, "", "  ")
	} else {
		result, e = json.Marshal(responseToWrite)
	}

	if e != nil {
		panic("error while generating JSON output")
	}
	request.Response().WriteHeader(code)
	request.Response().Write(result)
}

//Redirect redirects request to new url
func (request Request) Redirect(url string) {
	request.Header().Set("Location", url)
	request.Response().WriteHeader(http.StatusFound)
}

func (request Request) writeAfterLog() {
	if request.Request().Header.Get("X-Dont-Log") != "true" {
		duration := time.Now().Sub(request.receivedAt)
		request.Log().WithField("uuid", request.uuid).WithField("took", duration).
			Println(
				request.Request().Method,
				request.Request().URL.String(),
			)
	}
}

func (request Request) removeTrailingSlash() bool {
	path := request.Request().URL.Path
	if request.Request().Method == "GET" && len(path) > 1 && path == request.Request().URL.String() && strings.HasSuffix(path, "/") {
		request.Header().Set("Location", path[0:len(path)-1])
		request.Response().WriteHeader(http.StatusMovedPermanently)
		return true
	}
	return false
}

func parseRequest(r Request) {
	contentType := r.Request().Header.Get("Content-Type")
	var err error

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err = r.Request().ParseMultipartForm(1000000)
		if err != nil {
			panic(err)
		}

		for k, values := range r.Request().MultipartForm.Value {
			for _, v := range values {
				r.Request().Form.Add(k, v)
			}
		}
	} else {
		err = r.Request().ParseForm()
		if err != nil {
			panic(err)
		}
	}
}
