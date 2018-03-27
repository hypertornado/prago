package admin

import (
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"reflect"
)

type AdminNavigationPage struct {
	Navigation   AdminItemNavigation
	PageTemplate string
	PageData     interface{}
}

type AdminItemNavigation struct {
	Name        string
	PageTitle   string
	Tabs        []NavigationTab
	Breadcrumbs []NavigationBreadcrumb
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

func renderNavigationPage(request prago.Request, page AdminNavigationPage) {
	request.SetData("admin_title", page.Navigation.PageTitle)
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)
	prago.Render(request, 200, "admin_layout")
}

func (admin *Admin) getResourceNavigation(resource Resource, user User, code string) AdminItemNavigation {
	tabs := []NavigationTab{}

	if resource.CanCreate {
		tabs = append(tabs, NavigationTab{
			Name:     messages.Messages.Get(user.Locale, "admin_new"),
			URL:      admin.GetURL(&resource, "new"),
			Selected: trueIfEqual(code, "new"),
		})
	}

	for _, v := range resource.ResourceActions {
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
		{resource.Name(user.Locale), admin.GetURL(&resource, "")},
	}

	name := ""
	for _, v := range tabs {
		if v.Selected {
			name = v.Name
		}
	}

	return AdminItemNavigation{
		Name:        name,
		PageTitle:   name + " " + resource.Name(user.Locale),
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
	}

	for _, v := range resource.ResourceItemActions {
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

	return AdminItemNavigation{
		Name:        name,
		PageTitle:   name + " " + resource.Name(user.Locale),
		Tabs:        tabs,
		Breadcrumbs: breadcrumbs,
	}
}

func getItemName(item interface{}, locale string) string {
	if item == nil {
		return ""
	}

	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("Name")
	if field.IsValid() {
		return field.String()
	}
	return ""
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
