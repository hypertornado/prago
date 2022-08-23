package prago

import (
	"strings"
)

type ActionIface interface {
	getName(string) string
	getMethod() string
	getURL() string
	getController() *controller
	getConstraints() []func(map[string]string) bool
	getHandler() func(*Request)
	getPermission() Permission
	getURLToken() string
	returnIsPriority() bool
}

func (action *Action) getName(locale string) string {
	if action.name != nil {
		return action.name(locale)
	}
	return action.url
}

func (action *Action) getMethod() string {
	return action.method
}

func (action *Action) getURLToken() string {
	return action.url
}

func (action *Action) getURL() string {
	if strings.HasPrefix(action.url, "/") {
		panic("url can't start with / character")
	}

	var url string
	if action.resource == nil {
		url = action.app.getAdminURL(action.url)
	} else {
		resource := action.resource
		if action.isItemAction {
			if action.url != "" {
				url = resource.getData().getURL(":id/" + action.url)
			} else {
				url = resource.getData().getURL(":id")
			}
		} else {
			url = resource.getData().getURL(action.url)
		}
	}
	return url
}

func (action *Action) getController() *controller {
	var controller *controller
	if action.resource != nil {
		controller = action.resource.getData().getResourceControl()
	} else {
		controller = action.app.adminController
	}
	return controller
}

func (action *Action) getHandler() func(*Request) {
	return func(request *Request) {
		if !action.app.authorize(request.user, action.permission) {
			render403(request)
			return
		}
		if action.handler != nil {
			action.handler(request)
		} else {
			var data interface{}
			if action.dataSource != nil {
				data = action.dataSource(request)
			}
			renderPage(request, page{
				App:          action.app,
				Navigation:   action.getnavigation(request),
				PageTemplate: action.template,
				PageData:     data,
			})
		}
	}
}

func (action *Action) getConstraints() []func(map[string]string) bool {
	constraints := []func(map[string]string) bool{}
	if action.isItemAction {
		constraints = append(constraints, constraintInt("id"))
	}
	return constraints
}

func (action *Action) getPermission() Permission {
	return action.permission
}

func (action *Action) returnIsPriority() bool {
	return action.isPriority
}
