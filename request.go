package prago

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Request represents structure for http request
type Request struct {
	uuid       string
	receivedAt time.Time
	w          http.ResponseWriter
	r          *http.Request
	app        *App
	session    *requestSession

	notifications []*notificationView
}

func (request Request) getNotificationsData() string {
	if request.notifications == nil {
		return ""
	} else {
		b, err := json.Marshal(request.notifications)
		must(err)
		return string(b)
	}
}

// Request returns underlying http.Request
func (request Request) Request() *http.Request { return request.r }

// Response returns underlying http.ResponseWriter
func (request Request) Response() http.ResponseWriter { return request.w }

// Params returns url.Values of request
func (request Request) Params() url.Values {
	return request.Request().Form
}

func (request Request) Param(name string) string {
	return request.Request().Form.Get(name)
}

// UserID returns id of logged in user, returns 0 if no user is logged
func (request Request) UserID() int64 {

	if request.session == nil {
		return 0
	}

	if request.session.session == nil {
		return 0
	}

	if request.session.session.Values == nil {
		return 0
	}

	userID, ok := request.session.session.Values[userIDSessionName].(int64)
	if !ok {
		return 0
	}

	return userID
}

func (request *Request) getUser() *user {
	user := request.app.UsersResource.Query(request.r.Context()).ID(request.UserID())
	if user == nil {
		return nil
	}
	return user
}

func (request *Request) role() string {
	userID := request.UserID()
	data := request.app.userDataCacheGet(userID)
	if data == nil {
		return ""
	}
	return data.role
}

func (request *Request) Name() string {
	userID := request.UserID()
	data := request.app.userDataCacheGet(userID)
	if data == nil {
		return ""
	}
	return data.name
}

func (request *Request) Locale() string {
	userID := request.UserID()
	data := request.app.userDataCacheGet(userID)
	if data == nil {
		return localeFromRequest(request)
	}
	return data.locale
}

func (request Request) Authorize(permission Permission) bool {
	var logged bool
	if request.UserID() > 0 {
		logged = true
	}
	return request.app.authorize(logged, request.role(), permission)
}

// WriteHTML renders HTML view with HTTP code
func (request Request) WriteHTML(statusCode int, templateName string, data any) {
	request.Response().Header().Add("Content-Type", "text/html; charset=utf-8")
	request.writeSessionIfDirty()
	request.Response().WriteHeader(statusCode)
	must(
		request.app.templates.templates.ExecuteTemplate(
			request.Response(),
			templateName,
			data,
		),
	)
}

// WriteJSON renders JSON with HTTP code
func (request Request) WriteJSON(statusCode int, data interface{}) {
	request.Response().Header().Add("Content-type", "application/json")
	request.writeSessionIfDirty()

	pretty := false
	if request.Param("pretty") == "true" {
		pretty = true
	}

	var responseToWrite interface{}
	if statusCode >= 400 {
		responseToWrite = map[string]interface{}{"error": data, "errorCode": statusCode}
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
	request.Response().WriteHeader(statusCode)
	request.Response().Write(result)
}

// Redirect redirects request to new url
func (request Request) Redirect(url string) {
	request.Response().Header().Set("Location", url)
	request.writeSessionIfDirty()
	request.Response().WriteHeader(http.StatusFound)
}

func (request Request) writeAfterLog() {
	if request.Request().Header.Get("X-Dont-Log") != "true" {
		request.app.Log().accessln(
			fmt.Sprintf("id=%s %s %s took=%v",
				request.uuid,
				request.Request().Method,
				request.Request().URL.String(),
				time.Since(request.receivedAt),
			),
		)
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
