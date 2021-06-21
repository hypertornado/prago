package prago

import (
	"encoding/json"
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
	app        App
	session    *requestSession
	user       *user
}

//Request returns underlying http.Request
func (request Request) Request() *http.Request { return request.r }

//Response returns underlying http.ResponseWriter
func (request Request) Response() http.ResponseWriter { return request.w }

//Params returns url.Values of request
func (request Request) Params() url.Values {
	return request.Request().Form
}

//UserID returns id of logged in user, returns 0 if no user is logged
func (request Request) UserID() int64 {
	if request.user != nil {
		return request.user.ID
	}
	return 0
}

func (request Request) UserHasPermission(permission Permission) bool {
	if request.user == nil {
		return false
	}
	return request.app.authorize(request.user, permission)
}

//SetData sets request data
func (request Request) SetData(k string, v interface{}) { request.data[k] = v }

//GetData returns request data
func (request Request) GetData(k string) interface{} { return request.data[k] }

//RenderView with HTTP 200 code
func (request Request) RenderView(templateName string) {
	request.RenderViewWithCode(templateName, 200)
}

//RenderViewWithCode renders view with HTTP code
func (request Request) RenderViewWithCode(templateName string, statusCode int) {
	request.Response().Header().Add("Content-Type", "text/html; charset=utf-8")
	request.writeSessionIfDirty()
	request.Response().WriteHeader(statusCode)
	must(
		request.app.templates.templates.ExecuteTemplate(
			request.Response(),
			templateName,
			request.data,
		),
	)
}

//RenderJSON renders JSON with HTTP 200 code
func (request Request) RenderJSON(data interface{}) {
	request.RenderJSONWithCode(data, 200)
}

//RenderJSONWithCode renders JSON with HTTP code
func (request Request) RenderJSONWithCode(data interface{}, code int) {
	request.Response().Header().Add("Content-type", "application/json")
	request.writeSessionIfDirty()

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

	if pretty {
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
	request.Response().Header().Set("Location", url)
	request.writeSessionIfDirty()
	request.Response().WriteHeader(http.StatusFound)
}

func (request Request) writeAfterLog() {
	if request.Request().Header.Get("X-Dont-Log") != "true" {
		request.app.Log().Printf("id=%s %s %s took=%v", request.uuid,
			request.Request().Method, request.Request().URL.String(),
			time.Since(request.receivedAt))
	}
}

func (request Request) removeTrailingSlash() bool {
	path := request.Request().URL.Path
	if request.Request().Method == "GET" && len(path) > 1 && path == request.Request().URL.String() && strings.HasSuffix(path, "/") {
		request.Response().Header().Set("Location", path[0:len(path)-1])
		request.Response().WriteHeader(http.StatusMovedPermanently)
		return true
	}
	return false
}

func parseRequest(r *Request) {
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
