package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type API struct {
	app         *App
	method      string
	url         string
	permission  Permission
	resource    *Resource
	handler     func(*Request)
	handlerJSON func(*Request) interface{}
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

func APIJSON[T any](app *App, url string, handler func(*Request, *T) any) *API {
	api := newAPI(app, url)

	api.handler = func(request *Request) {
		data, err := io.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}

		var item T
		err = json.Unmarshal(data, &item)
		if err != nil {
			panic(err)
		}

		retData := handler(request, &item)
		request.WriteJSON(200, retData)
	}

	return api
}

func ResourceAPI[T any](app *App, url string) *API {
	resource := getResource[T](app)
	return resource.api(url)
}

func (resource *Resource) api(url string) *API {
	api := newAPI(resource.app, url)
	api.resource = resource
	api.permission = resource.canView
	return api
}

func (api *API) Method(method string) *API {
	if !isHTTPMethodValid(method) {
		panic("unsupported method for API action: " + method)
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

// TODO: Deprecated
func (api *API) HandlerJSON(handler func(*Request) interface{}) *API {
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

	controller := app.adminController
	controller.routeHandler("ANY", app.getAdminURL("api/*"), renderAPINotFound)
}

func (api *API) initAPI() error {

	var controller *controller
	if api.resource != nil {
		controller = api.resource.getResourceControl()
	} else {
		controller = api.app.accessController
	}

	var url string
	if api.resource == nil {
		url = api.app.getAdminURL("api/" + api.url)
	} else {
		url = api.resource.getURL("api/" + api.url)
	}

	if api.handler == nil && api.handlerJSON == nil {
		return errors.New("no handler for API set")
	}

	var fn = func(request *Request) {
		if !request.Authorize(api.permission) {
			renderAPINotAuthorized(request)
			return
		}
		if api.handlerJSON != nil {
			data := api.handlerJSON(request)
			if !request.Written {
				request.WriteJSON(200, data)
			}
			return
		}
		if api.handler != nil {
			api.handler(request)
			return
		}

	}

	if api.permission == "" {
		return fmt.Errorf("Permission for api '%s %s' should not be empty", api.method, url)
	}

	controller.routeHandler(api.method, url, fn)

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
