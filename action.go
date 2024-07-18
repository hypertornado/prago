package prago

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
)

type buttonData struct {
	Icon     string
	Name     string
	URL      string
	Params   map[string]string
	Priority int64
}

// Action represents action
type Action struct {
	name          func(string) string
	icon          string
	permission    Permission
	method        string
	url           string
	handler       func(*Request)
	constraints   []routerConstraint
	parentBoard   *Board
	isPartOfBoard *Board

	app          *App
	resource     *Resource
	isItemAction bool
	isUserMenu   bool
	priority     int64

	childAction *Action
}

func initAction(action *Action) error {
	url := action.getURL()
	controller := action.getController()

	if action.permission == "" {
		panic(fmt.Sprintf("Permission for action '%s %s' should not be empty", action.method, url))
	}

	controller.routeHandler(action.method, url, action.handler, action.constraints...)
	return nil
}

func (app *App) initAllActions() {
	for _, v := range app.rootActions {
		err := initAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding root action %s %s: %s", v.method, v.name("en"), err))
		}
	}

	for _, resource := range app.resources {
		resource.initActions()
	}
}

func (resource *Resource) initActions() {
	for _, v := range resource.actions {
		err := initAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resource.id, v.method, v.name("en"), err))
		}
	}
	for _, v := range resource.itemActions {
		err := initAction(v)
		if err != nil {
			panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resource.id, v.method, v.name("en"), err))
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
		icon:        iconAction,
	}
}

func (app *App) Action(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

func ResourceAction[T any](app *App, url string) *Action {
	resource := getResource[T](app)
	return resource.action(url)
}

func (resource *Resource) action(url string) *Action {
	action := newAction(resource.app, url)
	action.resource = resource
	action.permission = resource.canView
	resource.actions = append(resource.actions, action)
	return action
}

func (action *Action) Name(name func(string) string) *Action {
	action.name = name
	return action
}

func (action *Action) Permission(permission Permission) *Action {
	must(action.app.validatePermission(permission))
	action.permission = permission
	if action.childAction != nil {
		action.childAction.permission = permission
	}
	return action
}

func (action *Action) Method(method string) *Action {
	if !isHTTPMethodValid(method) {
		panic("unsupported method for action: " + method)
	}
	action.method = method
	return action
}

func (action *Action) setPriority(priority int64) *Action {
	action.priority = priority
	return action
}

func (action *Action) Icon(icon string) *Action {
	action.icon = icon
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

func (action *Action) addConstraint(constraint routerConstraint) {
	action.constraints = append(action.constraints, constraint)
}

func (resource *Resource) getItemButtonData(userData UserData, item interface{}) (ret []*buttonData) {
	for _, v := range resource.itemActions {
		if v.method != "GET" {
			continue
		}
		if !userData.Authorize(v.permission) {
			continue
		}
		if v.url == "" {
			continue
		}
		name := v.name(userData.Locale())
		if v.url == "" {
			name = resource.previewer(userData, item).Name()
		}
		ret = append(ret, &buttonData{
			Icon:     v.icon,
			Name:     name,
			URL:      resource.getItemURL(item, v.url, userData),
			Priority: v.priority,
		},
		)
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Priority > ret[j].Priority {
			return true
		} else {
			return false
		}
	})
	return ret
}

func (resource *Resource) getListItemActions(userData UserData, item any, id int64) listItemActions {
	ret := listItemActions{
		MenuButtons: resource.getItemButtonData(userData, item),
	}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Icon: iconView,
		URL:  resource.getURL(fmt.Sprintf("%d", id)),
	})

	if userData.Authorize(resource.canUpdate) && resource.orderField != nil {
		ret.ShowOrderButton = true
	}

	return ret
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
				url = resource.getURL(":id/" + action.url)
			} else {
				url = resource.getURL(":id")
			}
		} else {
			url = resource.getURL(action.url)
		}
	}
	return url
}

func (action *Action) getController() *controller {
	if action.resource != nil {
		return action.resource.getResourceControl()
	} else {
		return action.app.adminController
	}
}

func (action *Action) Content(dataSource func(*Request) template.HTML) *Action {
	return action.ui(func(request *Request, pd *pageData) {
		pd.PageContent = dataSource(request)
	})
}

func (action *Action) uiView(templates *PragoTemplates, myTemplate string, dataSource func(*Request) any) *Action {
	return action.ui(func(request *Request, pd *pageData) {
		var data any
		if dataSource != nil {
			data = dataSource(request)
		}
		pd.PageContent = templates.ExecuteToHTML(myTemplate, data)
	})

}

func (action *Action) ui(uiHandler func(*Request, *pageData)) *Action {
	return action.Handler(func(request *Request) {
		pageData := createPageData(request)

		if action.isItemAction {
			item := action.resource.query(request.r.Context()).ID(request.Param("id"))
			pageData.Menu = action.app.getMenu(request, item)
		}

		uiHandler(request, pageData)
		pageData.renderPage(request)
	})
}

func (action *Action) Handler(handler func(*Request)) *Action {
	action.handler = func(request *Request) {
		if !request.Authorize(action.permission) {
			renderErrorPage(request, 403)
			return
		}
		handler(request)
	}
	return action
}
