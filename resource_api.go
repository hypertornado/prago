package prago

import (
	"errors"
	"fmt"
	"strings"
)

func initResourceAPIs(resource *Resource) {
	/*resource.AddAPI("test").HandlerJSON(
		func(request Request) interface{} {
			return "ok"
		},
	)*/
}

type API struct {
	app         *App
	method      string
	url         string
	permission  Permission
	resource    *Resource
	handler     func(Request)
	handlerJSON func(Request) interface{}
}

func newAPI(app *App, url string) *API {
	api := &API{
		app:    app,
		method: "GET",
		url:    url,
	}
	app.apis = append(app.apis, api)
	return api
}

func (app *App) AddAPI(url string) *API {
	api := newAPI(app, url)
	return api
}

func (resource *Resource) AddAPI(url string) *API {
	api := newAPI(resource.app, url)
	api.resource = resource
	api.permission = resource.CanView
	return api
}

func (api *API) Method(method string) *API {
	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		panic("unsupported method for action: " + method)
	}
	api.method = method
	return api
}

func (api *API) Permission(permission Permission) *API {
	api.permission = permission
	return api
}

func (api *API) Handler(handler func(Request)) *API {
	api.handler = handler
	return api
}

func (api *API) HandlerJSON(handler func(Request) interface{}) *API {
	api.handlerJSON = handler
	return api
}

func (app *App) initAPIs() {
	for _, v := range app.apis {
		err := v.initAPI()
		if err != nil {
			panic(fmt.Sprintf("error while initializing api %s: %s", v.url, err))
		}
	}

	//TODO: support ANY
	app.AdminController.Get(app.GetAdminURL("api/*"), renderAPINotFound)
	app.AdminController.Post(app.GetAdminURL("api/*"), renderAPINotFound)
	app.AdminController.Delete(app.GetAdminURL("api/*"), renderAPINotFound)
	app.AdminController.Put(app.GetAdminURL("api/*"), renderAPINotFound)
}

func (api *API) initAPI() error {
	var controller *Controller
	if api.resource != nil {
		controller = api.resource.resourceController
	} else {
		controller = api.app.AdminController
	}

	var url string
	if api.resource == nil {
		url = api.app.GetAdminURL("api/" + api.url)
	} else {
		url = api.resource.getURL("api/" + api.url)
	}

	if api.handler == nil && api.handlerJSON == nil {
		return errors.New("no handler for API set")
	}

	var fn = func(request Request) {
		user := request.GetUser()
		if !api.app.Authorize(user, api.permission) {
			renderAPINotAuthorized(request)
			return
		}
		if api.handlerJSON != nil {
			data := api.handlerJSON(request)
			request.RenderJSON(data)
			return
		}
		if api.handler != nil {
			api.handler(request)
			return
		}
	}

	switch api.method {
	case "POST":
		controller.Post(url, fn)
	case "GET":
		controller.Get(url, fn)
	case "PUT":
		controller.Put(url, fn)
	case "DELETE":
		controller.Delete(url, fn)
	default:
		return fmt.Errorf("unknown method %s", api.method)
	}
	return nil
}

func renderAPINotAuthorized(request Request) {
	renderAPICode(request, 403)
}

func renderAPINotFound(request Request) {
	renderAPICode(request, 404)
}

func renderAPICode(request Request, code int) {
	var message string
	switch code {
	case 403:
		message = "Forbidden"
	case 404:
		message = "Not found"
	}

	request.Response().WriteHeader(code)
	request.Response().Write([]byte(fmt.Sprintf("%d - %s", code, message)))

}
