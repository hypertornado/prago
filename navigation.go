package prago

import (
	"strconv"

	"github.com/hypertornado/prago/messages"
)

type adminNavigationPage struct {
	Name         string
	App          *App
	Navigation   adminItemNavigation
	PageTemplate string
	PageData     interface{}
	HideBox      bool
}

type adminItemNavigation struct {
	Tabs []navigationTab
	Wide bool
}

type navigationTab struct {
	Name     string
	URL      string
	Selected bool
}

func isTabVisible(tabs []navigationTab, pos int) bool {
	if tabs[pos-1].Selected {
		return false
	}
	if tabs[pos].Selected {
		return false
	}
	return true
}

func renderNavigationPage(request Request, page adminNavigationPage) {
	renderNavigation(request, page, "admin_layout")
}

func renderNavigationPageNoLogin(request Request, page adminNavigationPage) {
	renderNavigation(request, page, "admin_layout_nologin")
}

func renderNavigation(request Request, page adminNavigationPage, viewName string) {

	var name string
	name = page.Name
	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			name = v.Name
		}
	}

	if name == "" {
		mainMenu := request.GetData("main_menu").(mainMenu)
		name = mainMenu.GetTitle()
	}

	request.SetData("admin_title", name)
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView(viewName)
}

func (app *App) getAdminNavigation(user User, code string) adminItemNavigation {
	tabs := []navigationTab{
		{
			Name:     messages.Messages.Get(user.Locale, "admin_signpost"),
			URL:      app.GetAdminURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range app.rootActions {
		if v.Method == "" || v.Method == "GET" {
			if app.Authorize(user, v.Permission) {
				tabs = append(tabs, navigationTab{
					Name:     v.getName(user.Locale),
					URL:      app.GetAdminURL(v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (resource Resource) getNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab
	for _, v := range resource.actions {
		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if resource.App.Authorize(user, v.Permission) {
				name := v.URL
				if v.Name != nil {
					name = v.Name(user.Locale)
				}
				tabs = append(tabs, navigationTab{
					Name:     name,
					URL:      resource.GetURL(v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (resource Resource) getItemNavigation(user User, item interface{}, code string) adminItemNavigation {
	var tabs []navigationTab
	for _, v := range resource.itemActions {
		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			name := v.URL
			if v.URL == "" {
				name = getItemName(item)
			}
			if v.Name != nil {
				name = v.Name(user.Locale)
			}
			if resource.App.Authorize(user, v.Permission) {
				tabs = append(tabs, navigationTab{
					Name:     name,
					URL:      resource.GetItemURL(item, v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (app *App) getSettingsNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_settings"),
		URL:      app.GetAdminURL("user/settings"),
		Selected: trueIfEqual(code, "settings"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_password_change"),
		URL:      app.GetAdminURL("user/password"),
		Selected: trueIfEqual(code, "password"),
	})

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (app *App) getNologinNavigation(language, code string) adminItemNavigation {
	tabs := []navigationTab{}

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_login_action"),
		URL:      app.GetAdminURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_register"),
		URL:      app.GetAdminURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_forgotten"),
		URL:      app.GetAdminURL("user/forgot"),
		Selected: trueIfEqual(code, "forgot"),
	})

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func trueIfEqual(a, b string) bool {
	if a == b {
		return true
	}
	return false
}

func createNavigationalItemHandler(action, templateName string, dataGenerator func(Resource, Request, User) interface{}) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		must(resource.App.Query().WhereIs("id", int64(id)).Get(item))

		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			App:          resource.App,
			Navigation:   resource.getItemNavigation(user, item, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalItemAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, Request, User) interface{}) Action {
	return Action{
		URL:     url,
		Name:    name,
		Handler: createNavigationalItemHandler(url, templateName, dataGenerator),
	}
}

func createNavigationalHandler(action, templateName string, dataGenerator func(Resource, Request, User) interface{}) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			App:          resource.App,
			Navigation:   resource.getNavigation(user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createNavigationalHandler(url, templateName, dataGenerator),
	}
}

func createAdminHandler(action, templateName string, dataGenerator func(Resource, Request, User) interface{}, empty bool) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		adminNavigation := resource.App.getAdminNavigation(user, action)
		var name string
		for _, v := range adminNavigation.Tabs {
			if v.Selected {
				name = v.Name
			}
		}

		renderNavigationPage(request, adminNavigationPage{
			Name:         name,
			App:          resource.App,
			PageTemplate: templateName,
			PageData:     data,
			HideBox:      empty,
		})
	}
}

func CreateAdminAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator, false),
	}
}

func CreateAdminEmptyAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator, true),
	}
}
