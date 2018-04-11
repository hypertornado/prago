package admin

import (
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
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

func (admin *Admin) getAdminNavigation(user User, code string) AdminItemNavigation {
	tabs := []NavigationTab{
		NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_admin"),
			URL:      admin.GetURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range admin.rootActions {
		if v.Auth(&user) {
			tabs = append(tabs, NavigationTab{
				Name:     v.GetName(user.Locale),
				URL:      admin.GetURL(v.Url),
				Selected: trueIfEqual(code, v.Url),
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

func (admin *Admin) getResourceNavigation(resource Resource, user User, code string) AdminItemNavigation {
	tabs := []NavigationTab{
		NavigationTab{
			Name:     resource.Name(user.Locale),
			URL:      resource.GetURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	if resource.CanCreate {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_new"),
			URL:      resource.GetURL("new"),
			Selected: trueIfEqual(code, "new"),
		})
	}

	if resource.CanExport {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_export"),
			URL:      resource.GetURL("export"),
			Selected: trueIfEqual(code, "export"),
		})
	}

	if resource.ActivityLog {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_history"),
			URL:      resource.GetURL("history"),
			Selected: trueIfEqual(code, "history"),
		})
	}

	for _, v := range resource.resourceActions {
		if v.Url == "" {
			continue
		}
		name := v.Url
		if v.Name != nil {
			name = v.Name(user.Locale)
		}

		if v.Auth == nil || v.Auth(&user) {
			tabs = append(tabs, NavigationTab{
				Name:     name,
				URL:      resource.GetURL(v.Url),
				Selected: trueIfEqual(code, v.Url),
			})
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
		breadcrumbs = append(breadcrumbs, NavigationBreadcrumb{resource.Name(user.Locale), resource.GetURL("")})
	} else {
		name = resource.Name(user.Locale)
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Admin) getItemNavigation(resource Resource, user User, item interface{}, code string) AdminItemNavigation {
	tabs := []NavigationTab{}
	name := getItemName(item, user.Locale)

	tabs = append(tabs, NavigationTab{
		Name:     name,
		URL:      resource.GetItemURL(item, ""),
		Selected: trueIfEqual(code, ""),
	})

	if resource.PreviewURLFunction != nil {
		url := resource.PreviewURLFunction(item)
		if url != "" {
			tabs = append(tabs, NavigationTab{
				Name: messages.Messages.Get(user.Locale, "admin_preview"),
				URL:  url,
			})
		}
	}

	if resource.CanEdit {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_edit"),
			URL:      resource.GetItemURL(item, "edit"),
			Selected: trueIfEqual(code, "edit"),
		})

		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_delete"),
			URL:      resource.GetItemURL(item, "delete"),
			Selected: trueIfEqual(code, "delete"),
		})
	}

	if resource.ActivityLog {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_history"),
			URL:      resource.GetItemURL(item, "history"),
			Selected: trueIfEqual(code, "history"),
		})
	}

	for _, v := range resource.resourceItemActions {
		if v.Name == nil {
			continue
		}
		name := v.Url
		if v.Name != nil {
			name = v.Name(user.Locale)
		}

		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if v.Auth == nil || v.Auth(&user) {
				tabs = append(tabs, NavigationTab{
					Name:     name,
					URL:      resource.GetItemURL(item, v.Url),
					Selected: trueIfEqual(code, v.Url),
				})
			}
		}
	}

	breadcrumbs := []NavigationBreadcrumb{
		{admin.HumanName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
		{resource.Name(user.Locale), resource.GetURL("")},
	}

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

func (admin *Admin) getSettingsNavigation(user User, code string) AdminItemNavigation {

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

func (admin *Admin) getNologinNavigation(language, code string) AdminItemNavigation {
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

func createNavigationalItemHandler(action, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) func(Admin, Resource, prago.Request, User) {
	return func(admin Admin, resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(admin, resource, request, user)
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalItemAction(url string, name func(string) string, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) Action {
	return Action{
		Url:     url,
		Name:    name,
		Handler: createNavigationalItemHandler(url, templateName, dataGenerator),
	}
}

func createNavigationalHandler(action, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) func(Admin, Resource, prago.Request, User) {
	return func(admin Admin, resource Resource, request prago.Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(admin, resource, request, user)
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(resource, user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateNavigationalAction(url string, name func(string) string, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		Url:     url,
		Handler: createNavigationalHandler(url, templateName, dataGenerator),
	}
}

func createAdminHandler(action, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) func(Admin, Resource, prago.Request, User) {
	return func(admin Admin, resource Resource, request prago.Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(admin, resource, request, user)
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getAdminNavigation(user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func CreateAdminAction(url string, name func(string) string, templateName string, dataGenerator func(Admin, Resource, prago.Request, User) interface{}) Action {
	return Action{
		Name:    name,
		Auth:    AuthenticateAdmin,
		Url:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator),
	}
}
