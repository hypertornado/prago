package prago

import (
	"strings"
)

type ActionIface interface {
	getIcon() string
	getName(string) string
	getMethod() string
	getURL() string
	getController() *controller
	getConstraints() []routerConstraint
	getHandler() func(*Request)
	getPermission() Permission
	getURLToken() string
	returnIsPriority() bool
}

func (action *Action) getIcon() string {
	return action.icon
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
	if action.resourceData == nil {
		url = action.app.getAdminURL(action.url)
	} else {
		resourceData := action.resourceData
		if action.isItemAction {
			if action.url != "" {
				url = resourceData.getURL(":id/" + action.url)
			} else {
				url = resourceData.getURL(":id")
			}
		} else {
			url = resourceData.getURL(action.url)
		}
	}
	return url
}

func (action *Action) getController() *controller {
	var controller *controller
	if action.resourceData != nil {
		controller = action.resourceData.getResourceControl()
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

func (action *Action) getConstraints() []routerConstraint {
	return action.constraints
}

func (action *Action) getPermission() Permission {
	return action.permission
}

func (action *Action) returnIsPriority() bool {
	return action.isPriority
}
