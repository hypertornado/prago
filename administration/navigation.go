package administration

import (
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
	"strconv"
	"strings"
)

type AdminNavigationPage struct {
	Navigation   AdminItemNavigation
	PageTemplate string
	PageData     interface{}
}

type AdminItemNavigation struct {
	Name        string
	Tabs        []NavigationTab
	Breadcrumbs []NavigationBreadcrumb
	Wide        bool
}

type NavigationTab struct {
	Name     string
	URL      string
	Selected bool
}

type NavigationBreadcrumb struct {
	Name string
	URL  string
}

func (navigation AdminItemNavigation) GetPageTitle() string {
	ret := []string{}
	for _, v := range navigation.Breadcrumbs {
		ret = append([]string{v.Name}, ret...)
	}
	ret = append([]string{navigation.Name}, ret...)
	return strings.Join(ret, " — ")
}

func renderNavigationPage(request prago.Request, page AdminNavigationPage) {
	request.SetData("admin_title", page.Navigation.GetPageTitle())
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView("admin_layout")
}

func renderNavigationPageNoLogin(request prago.Request, page AdminNavigationPage) {
	request.SetData("admin_title", page.Navigation.GetPageTitle())
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	request.RenderView("admin_layout_nologin")
}

func (admin *Administration) getAdminNavigation(user User, code string) AdminItemNavigation {
	tabs := []NavigationTab{
		NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_admin"),
			URL:      admin.GetURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range admin.rootActions {
		if admin.Authorize(user, v.Permission) {
			tabs = append(tabs, NavigationTab{
				Name:     v.GetName(user.Locale),
				URL:      admin.GetURL(v.URL),
				Selected: trueIfEqual(code, v.URL),
			})
		}
	}

	name := messages.Messages.Get(user.Locale, "admin_admin")

	breadcrumbs := []NavigationBreadcrumb{
		{admin.HumanName, "/"},
	}

	for _, v := range tabs {
		if v.Selected && v.URL != admin.Prefix {
			name = v.Name
			breadcrumbs = append(breadcrumbs, NavigationBreadcrumb{
				Name: messages.Messages.Get(user.Locale, "admin_admin"),
				URL:  admin.GetURL(""),
			})
		}
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Administration) getResourceNavigation(resource Resource, user User, code string) AdminItemNavigation {
	var tabs []NavigationTab
	for _, v := range resource.actions {
		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if admin.Authorize(user, v.Permission) {
				name := v.URL
				if v.Name != nil {
					name = v.Name(user.Locale)
				}
				tabs = append(tabs, NavigationTab{
					Name:     name,
					URL:      resource.GetURL(v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	breadcrumbs := []NavigationBreadcrumb{
		{admin.HumanName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
	}

	name := ""
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	if code != "" {
		breadcrumbs = append(breadcrumbs, NavigationBreadcrumb{resource.HumanName(user.Locale), resource.GetURL("")})
	} else {
		name = resource.HumanName(user.Locale)
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Administration) getItemNavigation(resource Resource, user User, item interface{}, code string) AdminItemNavigation {
	var tabs []NavigationTab
	for _, v := range resource.itemActions {
		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			name := v.URL
			if v.URL == "" {
				name = getItemName(item, user.Locale)
			}
			if v.Name != nil {
				name = v.Name(user.Locale)
			}
			if admin.Authorize(user, v.Permission) {
				tabs = append(tabs, NavigationTab{
					Name:     name,
					URL:      resource.GetItemURL(item, v.URL),
					Selected: trueIfEqual(code, v.URL),
				})
			}
		}
	}

	breadcrumbs := []NavigationBreadcrumb{
		{admin.HumanName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
		{resource.HumanName(user.Locale), resource.GetURL("")},
	}

	name := getItemName(item, user.Locale)
	if code != "" {
		breadcrumbs = append(breadcrumbs,
			NavigationBreadcrumb{name, resource.GetItemURL(item, "")})
		for _, v := range tabs {
			if v.Selected {
				name = v.Name
			}
		}
	}

	if code == "delete" {
		name = messages.Messages.Get(user.Locale, "admin_delete_confirmation")
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Administration) getSettingsNavigation(user User, code string) AdminItemNavigation {

	breadcrumbs := []NavigationBreadcrumb{
		{admin.HumanName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
	}

	tabs := []NavigationTab{}

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_settings"),
		URL:      admin.GetURL("user/settings"),
		Selected: trueIfEqual(code, "settings"),
	})

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_password_change"),
		URL:      admin.GetURL("user/password"),
		Selected: trueIfEqual(code, "password"),
	})

	if code == "password" {
		breadcrumbs = append(breadcrumbs,
			NavigationBreadcrumb{messages.Messages.Get(user.Locale, "admin_settings"), admin.GetURL("/user/settings")},
		)
	}

	var name string
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Administration) getNologinNavigation(language, code string) AdminItemNavigation {
	tabs := []NavigationTab{}

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(language, "admin_login_action"),
		URL:      admin.GetURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(language, "admin_register"),
		URL:      admin.GetURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, NavigationTab{
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

	return AdminItemNavigation{
		Name: name,
		Tabs: tabs,
		Breadcrumbs: []NavigationBreadcrumb{
			{admin.HumanName, "/"},
		},
	}
}

func getItemName(item interface{}, locale string) string {
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		field := itemsVal.FieldByName("Name")
		if field.IsValid() {
			ret := field.String()
			if ret != "" {
				return ret
			}
		}
	}
	return fmt.Sprintf("#%d", getItemID(item))
}

func getItemID(item interface{}) int64 {
	if item == nil {
		return 0
	}

	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("ID")
	if field.IsValid() {
		return field.Int()
	}
	return 0
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

		renderNavigationPage(request, AdminNavigationPage{
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

		renderNavigationPage(request, AdminNavigationPage{
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

		renderNavigationPage(request, AdminNavigationPage{
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
