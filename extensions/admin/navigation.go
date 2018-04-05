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
	prago.Render(request, 200, "admin_layout")
}

func renderNavigationPageNoLogin(request prago.Request, page AdminNavigationPage) {
	request.SetData("admin_title", page.Navigation.GetPageTitle())
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	prago.Render(request, 200, "admin_layout_nologin")
}

func (admin *Admin) getAdminNavigation(user User, code string) AdminItemNavigation {
	tabs := []NavigationTab{
		NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_admin"),
			URL:      admin.Prefix,
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range admin.rootActions {
		tabs = append(tabs, NavigationTab{
			Name:     v.GetName(user.Locale),
			URL:      admin.Prefix + "/" + v.Url,
			Selected: trueIfEqual(code, v.Url),
		})
	}

	name := messages.Messages.Get(user.Locale, "admin_admin")

	breadcrumbs := []NavigationBreadcrumb{
		{admin.AppName, "/"},
	}

	for _, v := range tabs {
		if v.Selected && v.URL != admin.Prefix {
			name = v.Name
			breadcrumbs = append(breadcrumbs, NavigationBreadcrumb{
				Name: messages.Messages.Get(user.Locale, "admin_admin"),
				URL:  admin.Prefix,
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
			Name:     messages.Messages.Get(user.Locale, "admin_list"),
			URL:      admin.GetURL(&resource, ""),
			Selected: trueIfEqual(code, ""),
		},
	}

	if resource.CanCreate {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_new"),
			URL:      admin.GetURL(&resource, "new"),
			Selected: trueIfEqual(code, "new"),
		})
	}

	if resource.ActivityLog {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_export"),
			URL:      admin.GetURL(&resource, "export"),
			Selected: trueIfEqual(code, "export"),
		})
	}

	if resource.ActivityLog {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_history"),
			URL:      admin.GetURL(&resource, "history"),
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
				URL:      admin.GetURL(&resource, v.Url),
				Selected: trueIfEqual(code, v.Url),
			})
		}
	}

	breadcrumbs := []NavigationBreadcrumb{
		{admin.AppName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
	}

	name := ""
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	if code != "" {
		breadcrumbs = append(breadcrumbs, NavigationBreadcrumb{resource.Name(user.Locale), admin.GetURL(&resource, "")})
	} else {
		name = resource.Name(user.Locale)
	}

	return AdminItemNavigation{
		Name:        name,
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func (admin *Admin) getItemNavigation(resource Resource, user User, item interface{}, itemID int, code string) AdminItemNavigation {
	prefix := admin.GetURL(&resource, fmt.Sprintf("%d", itemID))

	tabs := []NavigationTab{}

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_view"),
		URL:      prefix,
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
			URL:      prefix + "/edit",
			Selected: trueIfEqual(code, "edit"),
		})

		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_delete"),
			URL:      prefix + "/delete",
			Selected: trueIfEqual(code, "delete"),
		})
	}

	if resource.ActivityLog {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_history"),
			URL:      prefix + "/history",
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
					URL:      prefix + "/" + v.Url,
					Selected: trueIfEqual(code, v.Url),
				})
			}
		}
	}

	breadcrumbs := []NavigationBreadcrumb{
		{admin.AppName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
		{resource.Name(user.Locale), admin.GetURL(&resource, "")},
	}

	name := getItemName(item, user.Locale)

	if code != "" {
		breadcrumbs = append(breadcrumbs,
			NavigationBreadcrumb{name, prefix})
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
		{admin.AppName, "/"},
		{messages.Messages.Get(user.Locale, "admin_admin"), admin.Prefix},
	}

	tabs := []NavigationTab{}

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_settings"),
		URL:      admin.Prefix + "/user/settings",
		Selected: trueIfEqual(code, "settings"),
	})

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(user.Locale, "admin_password_change"),
		URL:      admin.Prefix + "/user/password",
		Selected: trueIfEqual(code, "password"),
	})

	if code == "password" {
		breadcrumbs = append(breadcrumbs,
			NavigationBreadcrumb{messages.Messages.Get(user.Locale, "admin_settings"), admin.Prefix + "/user/settings"},
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
		URL:      admin.Prefix + "/user/login",
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(language, "admin_register"),
		URL:      admin.Prefix + "/user/registration",
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, NavigationTab{
		Name:     messages.Messages.Get(language, "admin_forgotten"),
		URL:      admin.Prefix + "/user/forgot",
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
			{admin.AppName, "/"},
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
			Navigation:   admin.getItemNavigation(resource, user, item, id, action),
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
