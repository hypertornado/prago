package prago

import (
	"fmt"
	"strings"
)

type buttonData struct {
	Icon   string
	Name   string
	URL    string
	Params map[string]string
}

// Action represents action
type Action struct {
	name          func(string) string
	icon          string
	permission    Permission
	method        string
	url           string
	handler       func(*Request)
	template      string
	dataSource    func(*Request) interface{}
	constraints   []routerConstraint
	parentBoard   *Board
	isPartOfBoard *Board

	app            *App
	resourceData   *resourceData
	isItemAction   bool
	isUserMenu     bool
	isHiddenInMenu bool
	isPriority     bool
}

func bindAction(action *Action) error {
	url := action.getURL()
	controller := action.getController()

	var fn = action.getActionHandler()

	if action.permission == "" {
		panic(fmt.Sprintf("Permission for action '%s %s' should not be empty", action.method, url))
	}

	switch action.method {
	case "POST":
		controller.post(url, fn, action.constraints...)
	case "GET":
		controller.get(url, fn, action.constraints...)
	case "PUT":
		controller.put(url, fn, action.constraints...)
	case "DELETE":
		controller.delete(url, fn, action.constraints...)
	default:
		return fmt.Errorf("unknown method %s", action.method)
	}
	return nil
}

func (app *App) bindAllActions() {
	for _, v := range app.rootActions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding root action %s %s: %s", v.method, v.name("en"), err))
		}
	}

	for _, resourceData := range app.resources {
		resourceData.bindActions()
	}
}

func (resourceData *resourceData) bindActions() {
	for _, v := range resourceData.actions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resourceData.id, v.method, v.name("en"), err))
		}
	}
	for _, v := range resourceData.itemActions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resourceData.id, v.method, v.name("en"), err))
		}
	}
}

func newAction(app *App, url string) *Action {
	return &Action{
		name:        unlocalized(url),
		permission:  "",
		method:      "GET",
		url:         url,
		app:         app,
		parentBoard: app.MainBoard,
	}
}

// AddAction adds action to root
func (app *App) Action(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

func (resource *Resource[T]) Action(url string) *Action {
	return resource.data.action(url)
}

// AddAction adds action to resource
func (resourceData *resourceData) action(url string) *Action {
	action := newAction(resourceData.app, url)
	action.resourceData = resourceData
	action.permission = resourceData.canView
	resourceData.actions = append(resourceData.actions, action)
	return action
}

// Name sets action name
func (action *Action) Name(name func(string) string) *Action {
	action.name = name
	return action
}

// Permission sets action permission
func (action *Action) Permission(permission Permission) *Action {
	must(action.app.validatePermission(permission))
	action.permission = permission
	return action
}

// Method sets action method (GET, POST, PUT or DELETE)
func (action *Action) Method(method string) *Action {
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		panic("unsupported method for action: " + method)
	}
	action.method = method
	return action
}

func (action *Action) priority() *Action {
	action.isPriority = true
	return action
}

// Handler sets action handler
func (action *Action) Handler(handler func(*Request)) *Action {
	if action.template != "" {
		panic("can't set both action handler and template")
	}
	if action.dataSource != nil {
		panic("can't set both action handler and dataSource")
	}
	action.handler = handler
	return action
}

func (action *Action) Icon(icon string) *Action {
	action.icon = icon
	return action
}

// Template sets action template
func (action *Action) Template(template string) *Action {
	if action.handler != nil {
		panic("can't set both action handler and template")
	}
	action.template = template
	return action
}

// DataSource sets action data source, which is used to render template
func (action *Action) DataSource(dataSource func(*Request) interface{}) *Action {
	if action.handler != nil {
		panic("can't set both action handler and dataSource")
	}
	action.dataSource = dataSource
	return action
}

func (action *Action) Board(board *Board) *Action {
	action.parentBoard = board
	return action
}

func (action *Action) userMenu() *Action {
	action.isUserMenu = true
	return action
}

func (action *Action) hiddenInMenu() *Action {
	action.isHiddenInMenu = true
	return action
}

func (action *Action) getnavigation(request *Request) navigation {
	if action.resourceData != nil {
		return action.resourceData.getnavigation(action, request)
	}
	return navigation{}
}

func (action *Action) addConstraint(constraint routerConstraint) {
	action.constraints = append(action.constraints, constraint)
}

func (resourceData *resourceData) getnavigation(action *Action, request *Request) navigation {
	if resourceData == nil {
		return navigation{}
	}

	code := action.url
	if action.isItemAction {
		item := resourceData.query(request.r.Context()).ID(request.Param("id"))
		if item != nil {
			return resourceData.getItemNavigation(request, item, code)
		} else {
			return navigation{}
		}
	}
	return resourceData.getResourceNavigation(request, code)

}

func (resourceData *resourceData) getListItemActions(userData UserData, item any, id int64) listItemActions {
	ret := listItemActions{}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Icon: iconView,
		URL:  resourceData.getURL(fmt.Sprintf("%d", id)),
	})

	navigation := resourceData.getItemNavigation(userData, item, "")

	for _, v := range navigation.Tabs {
		if !v.Selected {
			ret.MenuButtons = append(ret.MenuButtons, buttonData{
				Icon: v.Icon,
				Name: v.Name,
				URL:  v.URL,
			})
		}
	}

	if userData.Authorize(resourceData.canUpdate) && resourceData.orderField != nil {
		ret.ShowOrderButton = true
	}

	return ret
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
	if action.resourceData != nil {
		return action.resourceData.getResourceControl()
	} else {
		return action.app.adminController
	}
}

func (action *Action) getActionHandler() func(*Request) {
	return func(request *Request) {
		if !request.Authorize(action.permission) {
			renderErrorPage(request, 403)
			return
		}
		if action.handler != nil {
			action.handler(request)
		} else {
			var data interface{}
			if action.dataSource != nil {
				data = action.dataSource(request)
			}

			pageData := createPageData(request)
			pageData.Navigation = action.getnavigation(request)
			pageData.PageTemplate = action.template
			pageData.PageData = data
			pageData.renderPage(request)
		}
	}
}
