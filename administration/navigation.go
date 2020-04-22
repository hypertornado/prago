package administration

import (
	"strconv"
	"strings"

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
	Name        string
	Tabs        []navigationTab
	Breadcrumbs []navigationBreadcrumb
	Wide        bool
}

type navigationTab struct {
	Name     string
	URL      string
	Selected bool
}

type navigationBreadcrumb struct {
	Name  string
	URL   string
	Logo  string
	Icon  string
	Image string
}

func (bc navigationBreadcrumb) NoImage() bool {
	if bc.Logo == "" && bc.Icon == "" && bc.Image == "" {
		return true
	}
	return false
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

func (admin *Administration) createBreadcrumbs(locale string) []navigationBreadcrumb {
	return []navigationBreadcrumb{
		{Name: admin.HumanName, URL: "/", Logo: admin.Logo},
		{Name: messages.Messages.Get(locale, "admin_admin"), URL: admin.Prefix},
	}
}

func (navigation adminItemNavigation) getPageTitle() string {
	ret := []string{}
	for _, v := range navigation.Breadcrumbs {
		ret = append([]string{v.Name}, ret...)
	}
	ret = append([]string{navigation.Name}, ret...)
	return strings.Join(ret, " — ")
}

func renderNavigationPage(request prago.Request, page adminNavigationPage) {
	request.SetData("admin_title", page.Navigation.getPageTitle())
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView("admin_layout")
}

func renderNavigationPageNoLogin(request prago.Request, page adminNavigationPage) {
	request.SetData("admin_title", page.Navigation.getPageTitle())
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView("admin_layout_nologin")
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
		if admin.Authorize(user, v.Permission) {
			tabs = append(tabs, navigationTab{
				Name:     v.getName(user.Locale),
				URL:      admin.GetURL(v.URL),
				Selected: trueIfEqual(code, v.URL),
			})
		}
	}

	name := messages.Messages.Get(user.Locale, "admin_signpost")

	breadcrumbs := admin.createBreadcrumbs(user.Locale)

	for _, v := range tabs {
		if v.Selected && v.URL != admin.Prefix {
			name = v.Name
		}
	}

	return adminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
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

	breadcrumbs := admin.createBreadcrumbs(user.Locale)

	name := ""
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	breadcrumbs = append(breadcrumbs, navigationBreadcrumb{Name: resource.HumanName(user.Locale), URL: resource.GetURL("")})

	return adminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
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

	breadcrumbs := admin.createBreadcrumbs(user.Locale)
	breadcrumbs = append(breadcrumbs,
		navigationBreadcrumb{
			Name: resource.HumanName(user.Locale),
			URL:  resource.GetURL(""),
		})

	name := getItemName(item)
	breadcrumbs = append(breadcrumbs,
		navigationBreadcrumb{Name: name, URL: resource.GetItemURL(item, ""), Image: admin.getItemImage(item)})
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	if code == "delete" {
		name = messages.Messages.Get(user.Locale, "admin_delete_confirmation")
	}

	return adminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Administration) getSettingsNavigation(user User, code string) adminItemNavigation {
	breadcrumbs := admin.createBreadcrumbs(user.Locale)
	breadcrumbs = append(breadcrumbs,
		navigationBreadcrumb{
			Name: messages.Messages.Get(user.Locale, "admin_settings"),
			URL:  admin.GetURL("user/settings"),
		},
	)

	tabs := []navigationTab{}

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

	var name string
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	return adminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
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

	var name string
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	breadcrumbs := admin.createBreadcrumbs(language)

	return adminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: []navigationBreadcrumb{breadcrumbs[0]},
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

func createAdminHandler(action, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) func(Resource, prago.Request, User) {
	return func(resource Resource, request prago.Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		request.SetData("admin_navigation_logo_selected", true)
		renderNavigationPage(request, adminNavigationPage{
			Admin:        resource.Admin,
			Navigation:   resource.Admin.getAdminNavigation(user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateAdminAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator),
	}
}
