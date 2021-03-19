package prago

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hypertornado/prago/utils"
)

type buttonData struct {
	Name   string
	URL    string
	Params map[string]string
}

//Action represents action
type Action struct {
	name       func(string) string
	permission Permission
	method     string
	url        string
	handler    func(*Request)
	template   string
	dataSource func(*Request) interface{}

	app          *App
	resource     *Resource
	isItemAction bool
	isWide       bool
	isUserMenu   bool
	isHiddenMenu bool
	isPriority   bool
}

func (app *App) bindAllActions() {
	for _, v := range app.rootActions {
		err := v.bindAction()
		if err != nil {
			panic(fmt.Sprintf("error while binding root action %s %s: %s", v.method, v.name("en"), err))
		}
	}

	for _, resource := range app.resources {
		for _, v := range resource.actions {
			err := v.bindAction()
			if err != nil {
				panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resource.id, v.method, v.name("en"), err))
			}
		}
		for _, v := range resource.itemActions {
			err := v.bindAction()
			if err != nil {
				panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resource.id, v.method, v.name("en"), err))
			}
		}

	}

}

func newAction(app *App, url string) *Action {
	return &Action{
		name:       Unlocalized(url),
		permission: sysadminPermission,
		method:     "GET",
		url:        url,
		app:        app,
	}
}

//AddAction adds action to root
func (app *App) Action(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

//AddAction adds action to resource
func (resource *Resource) Action(url string) *Action {
	action := newAction(resource.app, url)
	action.resource = resource
	action.permission = resource.canView
	resource.actions = append(resource.actions, action)
	return action
}

//AddItemAction adds action to resource item
func (resource *Resource) ItemAction(url string) *Action {
	action := newAction(resource.app, url)
	action.resource = resource
	action.isItemAction = true
	action.permission = resource.canView
	resource.itemActions = append(resource.itemActions, action)
	return action
}

//Name sets action name
func (action *Action) Name(name func(string) string) *Action {
	action.name = name
	return action
}

//Permission sets action permission
func (action *Action) Permission(permission Permission) *Action {
	must(action.app.validatePermission(permission))
	action.permission = permission
	return action
}

//Method sets action method (GET, POST, PUT or DELETE)
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

//Handler sets action handler
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

//Template sets action template
func (action *Action) Template(template string) *Action {
	if action.handler != nil {
		panic("can't set both action handler and template")
	}
	action.template = template
	return action
}

//DataSource sets action data source, which is used to render template
func (action *Action) DataSource(dataSource func(*Request) interface{}) *Action {
	if action.handler != nil {
		panic("can't set both action handler and dataSource")
	}
	action.dataSource = dataSource
	return action
}

//IsWide sets rendering to fill whole realestate of page, no box is rendered
func (action *Action) IsWide() *Action {
	action.isWide = true
	return action
}

func (action *Action) userMenu() *Action {
	action.isUserMenu = true
	return action
}

func (action *Action) hiddenMenu() *Action {
	action.isHiddenMenu = true
	return action
}

func (action *Action) getnavigation(request *Request) navigation {
	if action.resource != nil {
		code := action.url
		if action.isItemAction {
			var item interface{}
			action.resource.newItem(&item)
			must(action.resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			return action.resource.getItemNavigation(request.user, item, code)
		}
		return action.resource.getNavigation(request.user, code)
	}
	return navigation{}

}

func (action *Action) bindAction() error {
	app := action.app
	if strings.HasPrefix(action.url, "/") {
		return errors.New("url can't start with / character")
	}

	var url string
	if action.resource == nil {
		url = app.getAdminURL(action.url)
	} else {
		resource := action.resource
		if action.isItemAction {
			if action.url != "" {
				url = resource.getURL(":id/" + action.url)
			} else {
				url = resource.getURL(":id")
			}
		} else {
			url = resource.getURL(action.url)
		}
	}

	var controller *controller
	if action.resource != nil {
		controller = action.resource.resourceController
	} else {
		controller = app.adminController
	}

	var fn = func(request *Request) {
		if !app.authorize(request.user, action.permission) {
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
			var hideBox bool
			if action.isWide {
				hideBox = true
			}
			renderNavigationPage(request, page{
				App:          app,
				Navigation:   action.getnavigation(request),
				PageTemplate: action.template,
				PageData:     data,
				HideBox:      hideBox,
			})
		}
	}

	constraints := []func(map[string]string) bool{}
	if action.isItemAction {
		constraints = append(constraints, utils.ConstraintInt("id"))
	}

	switch action.method {
	case "POST":
		controller.post(url, fn, constraints...)
	case "GET":
		controller.get(url, fn, constraints...)
	case "PUT":
		controller.put(url, fn, constraints...)
	case "DELETE":
		controller.delete(url, fn, constraints...)
	default:
		return fmt.Errorf("unknown method %s", action.method)
	}
	return nil
}

func (resource *Resource) getResourceActionsButtonData(user *User, admin *App) (ret []buttonData) {
	navigation := resource.getNavigation(user, "")
	for _, v := range navigation.Tabs {
		ret = append(ret, buttonData{
			Name: v.Name,
			URL:  v.URL,
		})
	}
	return
}

func (app *App) getListItemActions(user *User, item interface{}, id int64, resource Resource) listItemActions {
	ret := listItemActions{}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Name: messages.Get(user.Locale, "admin_view"),
		URL:  resource.getURL(fmt.Sprintf("%d", id)),
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

	if app.authorize(user, resource.canEdit) && resource.orderColumnName != "" {
		ret.ShowOrderButton = true
	}

	return ret
}
