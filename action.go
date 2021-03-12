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
	handler    func(Request)
	template   string
	dataSource func(Request) interface{}

	app          *App
	resource     *Resource
	isItemAction bool
	isWide       bool
	isUserMenu   bool
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
				panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resource.ID, v.method, v.name("en"), err))
			}
		}
		for _, v := range resource.itemActions {
			err := v.bindAction()
			if err != nil {
				panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resource.ID, v.method, v.name("en"), err))
			}
		}

	}

}

func newAction(app *App, url string) *Action {
	return &Action{
		name:       Unlocalized(url),
		permission: "",
		method:     "GET",
		url:        url,
		app:        app,
	}
}

//AddAction adds action to root
func (app *App) AddAction(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

//AddAction adds action to resource
func (resource *Resource) AddAction(url string) *Action {
	action := newAction(resource.App, url)
	action.resource = resource
	action.permission = resource.CanView
	resource.actions = append(resource.actions, action)
	return action
}

//AddItemAction adds action to resource item
func (resource *Resource) AddItemAction(url string) *Action {
	action := newAction(resource.App, url)
	action.resource = resource
	action.isItemAction = true
	action.permission = resource.CanView
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

//Template sets action template
func (action *Action) Template(template string) *Action {
	action.template = template
	return action
}

//Handler sets action handler
func (action *Action) Handler(handler func(Request)) *Action {
	action.handler = handler
	return action
}

//DataSource sets action data source, which is used to render template
func (action *Action) DataSource(dataSource func(Request) interface{}) *Action {
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

func (action *Action) getnavigation(request Request) adminItemNavigation {
	if action.resource != nil {
		user := request.GetUser()
		code := action.url
		if action.isItemAction {
			var item interface{}
			action.resource.newItem(&item)
			must(action.resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			return action.resource.getItemNavigation(user, item, code)
		}
		return action.resource.getNavigation(user, code)
	}
	return adminItemNavigation{}

}

func (action *Action) bindAction() error {
	app := action.app
	if strings.HasPrefix(action.url, "/") {
		return errors.New("url can't start with / character")
	}

	var url string
	if action.resource == nil {
		url = app.GetAdminURL(action.url)
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

	var controller *Controller
	if action.resource != nil {
		controller = action.resource.ResourceController
	} else {
		controller = app.AdminController
	}

	var fn = func(request Request) {
		user := request.GetUser()
		if !app.Authorize(user, action.permission) {
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
			renderNavigationPage(request, adminNavigationPage{
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
		controller.Post(url, fn, constraints...)
	case "GET":
		controller.Get(url, fn, constraints...)
	case "PUT":
		controller.Put(url, fn, constraints...)
	case "DELETE":
		controller.Delete(url, fn, constraints...)
	default:
		return fmt.Errorf("unknown method %s", action.method)
	}
	return nil
}

func (resource *Resource) getResourceActionsButtonData(user User, admin *App) (ret []buttonData) {
	navigation := resource.getNavigation(user, "")
	for _, v := range navigation.Tabs {
		ret = append(ret, buttonData{
			Name: v.Name,
			URL:  v.URL,
		})
	}
	return
}

func (app *App) getListItemActions(user User, item interface{}, id int64, resource Resource) listItemActions {
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

	if app.Authorize(user, resource.CanEdit) && resource.OrderColumnName != "" {
		ret.ShowOrderButton = true
	}

	return ret
}
