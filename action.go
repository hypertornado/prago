package prago

import (
	"fmt"
	"strings"
)

type buttonData struct {
	Name   string
	URL    string
	Params map[string]string
}

// Action represents action
type Action struct {
	name       func(string) string
	permission Permission
	method     string
	url        string
	handler    func(*Request)
	template   string
	dataSource func(*Request) interface{}

	app                *App
	resource           resourceIface
	isItemAction       bool
	isUserMenu         bool
	isHiddenInMainMenu bool
	isPriority         bool
}

func bindAction(action ActionIface) error {
	url := action.getURL()
	controller := action.getController()

	var fn = action.getHandler()
	constraints := action.getConstraints()

	if action.getPermission() == "" {
		panic(fmt.Sprintf("Permission for action '%s %s' should not be empty", action.getMethod(), url))
	}

	switch action.getMethod() {
	case "POST":
		controller.post(url, fn, constraints...)
	case "GET":
		controller.get(url, fn, constraints...)
	case "PUT":
		controller.put(url, fn, constraints...)
	case "DELETE":
		controller.delete(url, fn, constraints...)
	default:
		return fmt.Errorf("unknown method %s", action.getMethod())
	}
	return nil
}

func (app *App) bindAllActions() {
	for _, v := range app.rootActions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding root action %s %s: %s", v.getMethod(), v.getName("en"), err))
		}
	}

	for _, resource := range app.resources {
		resource.getData().bindActions()
	}
}

func (resourceData *resourceData) bindActions() {
	for _, v := range resourceData.actions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resourceData.id, v.getMethod(), v.getName("en"), err))
		}
	}
	for _, v := range resourceData.itemActions {
		err := bindAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resourceData.id, v.getMethod(), v.getName("en"), err))
		}
	}
}

func newAction(app *App, url string) *Action {
	return &Action{
		name:       unlocalized(url),
		permission: "",
		method:     "GET",
		url:        url,
		app:        app,
	}
}

// AddAction adds action to root
func (app *App) Action(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

// AddAction adds action to resource
func (resource *Resource[T]) Action(url string) *Action {
	action := newAction(resource.data.app, url)
	action.resource = resource
	action.permission = resource.data.canView
	resource.data.actions = append(resource.data.actions, action)
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
	method = strings.ToUpper(method)
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

func (action *Action) userMenu() *Action {
	action.isUserMenu = true
	return action
}

func (action *Action) hiddenInMainMenu() *Action {
	action.isHiddenInMainMenu = true
	return action
}

func (action *Action) getnavigation(request *Request) navigation {
	if action.resource != nil {
		return action.resource.getnavigation(action, request)
	}
	return navigation{}
}

func (resource *Resource[T]) getnavigation(action *Action, request *Request) navigation {
	if resource == nil {
		return navigation{}
	}

	code := action.url
	if action.isItemAction {
		item := resource.ID(request.Param("id"))
		if item != nil {
			return resource.getItemNavigation(request.user, item, code)
		} else {
			return navigation{}
		}
	}
	return resource.data.getResourceNavigation(request.user, code)

}

func (resource *Resource[T]) getListItemActions(user *user, item *T, id int64) listItemActions {
	ret := listItemActions{}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Name: messages.Get(user.Locale, "admin_view"),
		URL:  resource.getData().getURL(fmt.Sprintf("%d", id)),
	})

	navigation := resource.getItemNavigation(user, item, "")

	for _, v := range navigation.Tabs {
		if !v.Selected {
			ret.MenuButtons = append(ret.MenuButtons, buttonData{
				Name: v.Name,
				URL:  v.URL,
			})
		}
	}

	if resource.data.app.authorize(user, resource.data.canUpdate) && resource.data.orderField != nil {
		ret.ShowOrderButton = true
	}

	return ret
}
