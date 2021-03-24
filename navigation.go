package prago

type page struct {
	Name         string
	App          *App
	Navigation   navigation
	PageTemplate string
	PageData     interface{}
	HideBox      bool
}

type navigation struct {
	Tabs []tab
	Wide bool
}

type tab struct {
	Name     string
	URL      string
	Selected bool
	priority bool
}

func (n navigation) sortByPriority() navigation {
	var priorityTabs, nonPriorityTabs []tab
	for _, v := range n.Tabs {
		if v.priority {
			priorityTabs = append(priorityTabs, v)
		} else {
			nonPriorityTabs = append(nonPriorityTabs, v)
		}
	}
	n.Tabs = append(priorityTabs, nonPriorityTabs...)
	return n
}

func (nav page) Logo() string {
	return nav.App.logo
}

func isTabVisible(tabs []tab, pos int) bool {
	if tabs[pos-1].Selected {
		return false
	}
	if tabs[pos].Selected {
		return false
	}
	return true
}

func renderNavigationPage(request *Request, page page) {
	renderNavigation(request, page, "admin_layout")
}

func renderNavigationPageNoLogin(request *Request, page page) {
	renderNavigation(request, page, "admin_layout_nologin")
}

func renderNavigation(request *Request, page page, viewName string) {
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

func (resource Resource) getNavigation(user *user, code string) navigation {
	var tabs []tab
	for _, v := range resource.actions {
		if v.method == "GET" {
			if resource.app.authorize(user, v.permission) {
				name := v.url
				if v.name != nil {
					name = v.name(user.Locale)
				}
				tabs = append(tabs, tab{
					Name:     name,
					URL:      resource.getURL(v.url),
					Selected: trueIfEqual(code, v.url),
					priority: v.isPriority,
				})
			}
		}
	}

	return navigation{
		Tabs: tabs,
	}.sortByPriority()
}

func (resource Resource) getItemNavigation(user *user, item interface{}, code string) navigation {
	var tabs []tab
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
			if resource.app.authorize(user, v.permission) {
				tabs = append(tabs, tab{
					Name:     name,
					URL:      resource.getItemURL(item, v.url),
					Selected: trueIfEqual(code, v.url),
					priority: v.isPriority,
				})
			}
		}
	}

	return navigation{
		Tabs: tabs,
	}.sortByPriority()
}

func (app *App) getNologinNavigation(language, code string) navigation {
	tabs := []tab{}

	tabs = append(tabs, tab{
		Name:     messages.Get(language, "admin_login_action"),
		URL:      app.getAdminURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, tab{
		Name:     messages.Get(language, "admin_register"),
		URL:      app.getAdminURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, tab{
		Name:     messages.Get(language, "admin_forgotten"),
		URL:      app.getAdminURL("user/forgot"),
		Selected: trueIfEqual(code, "forgot"),
	})

	return navigation{
		Tabs: tabs,
	}.sortByPriority()
}

func trueIfEqual(a, b string) bool {
	if a == b {
		return true
	}
	return false
}
