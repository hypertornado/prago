package administration

import (
	"strconv"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
)

type adminNavigationPage struct {
	Admin        *Administration
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

func IsTabVisible(tabs []navigationTab, pos int) bool {
	if tabs[pos-1].Selected {
		return false
	}
	if tabs[pos].Selected {
		return false
	}
	return true
}

func renderNavigationPage(request prago.Request, page adminNavigationPage) {
	renderNavigation(request, page, "admin_layout")
}

func renderNavigationPageNoLogin(request prago.Request, page adminNavigationPage) {
	renderNavigation(request, page, "admin_layout_nologin")
}

func renderNavigation(request prago.Request, page adminNavigationPage, viewName string) {

	var name string
	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			name = v.Name
		}
	}

	request.SetData("admin_title", name)
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView(viewName)
}

func (admin *Administration) getAdminNavigation(user User, code string) adminItemNavigation {
	tabs := []navigationTab{
		{
			Name:     messages.Messages.Get(user.Locale, "admin_signpost"),
			URL:      admin.GetURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range admin.rootActions {
		if v.Method == "" || v.Method == "GET" {
			if admin.Authorize(user, v.Permission) {
				tabs = append(tabs, navigationTab{
					Name:     v.getName(user.Locale),
					URL:      admin.GetURL(v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (admin *Administration) getResourceNavigation(resource Resource, user User, code string) adminItemNavigation {
	var tabs []navigationTab
	for _, v := range resource.actions {
		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if admin.Authorize(user, v.Permission) {
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

func (admin *Administration) getItemNavigation(resource Resource, user User, item interface{}, code string) adminItemNavigation {
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
			if admin.Authorize(user, v.Permission) {
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

func (admin *Administration) getSettingsNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_settings"),
		URL:      admin.GetURL("user/settings"),
		Selected: trueIfEqual(code, "settings"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_password_change"),
		URL:      admin.GetURL("user/password"),
		Selected: trueIfEqual(code, "password"),
	})

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (admin *Administration) getNologinNavigation(language, code string) adminItemNavigation {
	tabs := []navigationTab{}

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_login_action"),
		URL:      admin.GetURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_register"),
		URL:      admin.GetURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Messages.Get(language, "admin_forgotten"),
		URL:      admin.GetURL("user/forgot"),
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

func createNavigationalItemHandler(action, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) func(Resource, prago.Request, User) {
	return func(resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))

		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			Admin:        resource.Admin,
			Navigation:   resource.Admin.getItemNavigation(resource, user, item, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalItemAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) Action {
	return Action{
		URL:     url,
		Name:    name,
		Handler: createNavigationalItemHandler(url, templateName, dataGenerator),
	}
}

func createNavigationalHandler(action, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) func(Resource, prago.Request, User) {
	return func(resource Resource, request prago.Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			Admin:        resource.Admin,
			Navigation:   resource.Admin.getResourceNavigation(resource, user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createNavigationalHandler(url, templateName, dataGenerator),
	}
}

func createAdminHandler(action, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}, empty bool) func(Resource, prago.Request, User) {
	return func(resource Resource, request prago.Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			Admin:        resource.Admin,
			Navigation:   resource.Admin.getAdminNavigation(user, action),
			PageTemplate: templateName,
			PageData:     data,
			HideBox:      empty,
		})
	}
}

func CreateAdminAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator, false),
	}
}

func CreateAdminEmptyAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator, true),
	}
}
