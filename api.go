package prago

import (
	"errors"
	"fmt"
	"strings"
)

type API struct {
	app          *App
	method       string
	url          string
	permission   Permission
	resourceData *resourceData
	handler      func(*Request)
	handlerJSON  func(*Request) interface{}
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

func (app *App) API(url string) *API {
	api := newAPI(app, url)
	return api
}

func (resource *Resource[T]) API(url string) *API {
	return resource.data.API(url)
}

func (resourceData *resourceData) API(url string) *API {
	api := newAPI(resourceData.app, url)
	api.resourceData = resourceData
	api.permission = resourceData.canView
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
	must(api.app.validatePermission(permission))
	api.permission = permission
	return api
}

func (api *API) Handler(handler func(*Request)) *API {
	api.handler = handler
	return api
}

func (api *API) HandlerJSON(handler func(*Request) interface{}) *API {
	api.handlerJSON = handler
	return api
}

func (app *App) bindAPIs() {
	for _, v := range app.apis {
		err := v.bindAPI()
		if err != nil {
			panic(fmt.Sprintf("error while initializing api %s: %s", v.url, err))
		}
	}

	controller := app.adminController

	//TODO: support ANY
	controller.get(app.getAdminURL("api/*"), renderAPINotFound)
	controller.post(app.getAdminURL("api/*"), renderAPINotFound)
	controller.delete(app.getAdminURL("api/*"), renderAPINotFound)
	controller.put(app.getAdminURL("api/*"), renderAPINotFound)
}

func (api *API) bindAPI() error {

	var controller *controller
	if api.resourceData != nil {
		controller = api.resourceData.getResourceControl()
	} else {
		//controller = api.app.adminController
		controller = api.app.accessController
	}

	var url string
	if api.resourceData == nil {
		url = api.app.getAdminURL("api/" + api.url)
	} else {
		url = api.resourceData.getURL("api/" + api.url)
	}

	if api.handler == nil && api.handlerJSON == nil {
		return errors.New("no handler for API set")
	}

	var fn = func(request *Request) {
		if !api.app.authorize(request.user, api.permission) {
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

	if api.permission == "" {
		panic(fmt.Sprintf("Permission for api '%s %s' should not be empty", api.method, url))
	}

	switch api.method {
	case "POST":
		controller.post(url, fn)
	case "GET":
		controller.get(url, fn)
	case "PUT":
		controller.put(url, fn)
	case "DELETE":
		controller.delete(url, fn)
	default:
		return fmt.Errorf("unknown method %s", api.method)
	}
	return nil
}

func renderAPINotAuthorized(request *Request) {
	renderAPICode(request, 403)
}

func renderAPINotFound(request *Request) {
	renderAPICode(request, 404)
}

func renderAPICode(request *Request, code int) {
	var message string
	switch code {
	case 403:
		message = "Forbidden"
	case 404:
		message = "Not found"
	}
	renderAPIMessage(request, code, message)
}

func renderAPIMessage(request *Request, code int, message string) {
	request.Response().WriteHeader(code)
	request.Response().Write([]byte(fmt.Sprintf("%d - %s", code, message)))

}
