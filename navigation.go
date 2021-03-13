package prago

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

func (resource Resource) getNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab
	for _, v := range resource.actions {
		if v.method == "GET" {
			if resource.app.Authorize(user, v.permission) {
				name := v.url
				if v.name != nil {
					name = v.name(user.Locale)
				}
				tabs = append(tabs, navigationTab{
					Name:     name,
					URL:      resource.getURL(v.url),
					Selected: trueIfEqual(code, v.url),
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
		if v.method == "GET" {
			name := v.url
			if v.url == "" {
				name = getItemName(item)
			} else {
				if v.name != nil {
					name = v.name(user.Locale)
				}
			}
			if resource.app.Authorize(user, v.permission) {
				tabs = append(tabs, navigationTab{
					Name:     name,
					URL:      resource.GetItemURL(item, v.url),
					Selected: trueIfEqual(code, v.url),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (app *App) getNologinNavigation(language, code string) adminItemNavigation {
	tabs := []navigationTab{}

	tabs = append(tabs, navigationTab{
		Name:     messages.Get(language, "admin_login_action"),
		URL:      app.GetAdminURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Get(language, "admin_register"),
		URL:      app.GetAdminURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Get(language, "admin_forgotten"),
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
