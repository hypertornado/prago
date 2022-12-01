package prago

type page struct {
	Name         string
	App          *App
	Navigation   navigation
	PageTemplate string
	PageData     interface{}
	HTTPCode     int
}

type navigation struct {
	Tabs []tab
	Wide bool
}

type tab struct {
	Icon     string
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
	return ""
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

func renderPage(request *Request, page page) {
	var name string
	name = page.Name
	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			name = v.Name
		}
	}

	if name == "" {
		mainMenu, ok := request.GetData("main_menu").(mainMenu)
		if ok {
			name = mainMenu.GetTitle()
		}
	}

	if request.app.logo != nil {
		request.SetData("admin_has_logo", true)
	}

	request.SetData("admin_title", name)
	request.SetData("admin_yield", "admin_navigation_page")
	request.SetData("admin_page", page)

	code := page.HTTPCode
	if page.HTTPCode == 0 {
		code = 200
	}

	layout := "admin_layout"
	if request.user == nil {
		request.SetData("language", localeFromRequest(request))
		layout = "admin_layout_nologin"
	}

	request.RenderViewWithCode(layout, code)
}

func (resourceData *resourceData) getResourceNavigation(user *user, code string) navigation {
	var tabs []tab
	for _, v := range resourceData.actions {
		if v.getMethod() == "GET" {
			if resourceData.app.authorize(user, v.getPermission()) {
				tabs = append(tabs, tab{
					Icon:     v.getIcon(),
					Name:     v.getName(user.Locale),
					URL:      resourceData.getURL(v.getURLToken()),
					Selected: trueIfEqual(code, v.getURLToken()),
					priority: v.returnIsPriority(),
				})
			}
		}
	}

	return navigation{
		Tabs: tabs,
	}.sortByPriority()
}

func (resourceData *resourceData) getItemNavigation(user *user, item interface{}, code string) navigation {
	var tabs []tab
	for _, v := range resourceData.itemActions {
		if v.getMethod() == "GET" {
			name := v.getName(user.Locale)
			if v.getURLToken() == "" {
				name = resourceData.previewer(user, item).Name()
			}
			if resourceData.app.authorize(user, v.getPermission()) {
				tabs = append(tabs, tab{
					Icon:     v.getIcon(),
					Name:     name,
					URL:      resourceData.getItemURL(item, v.getURLToken(), user),
					Selected: trueIfEqual(code, v.getURLToken()),
					priority: v.returnIsPriority(),
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
	} else {
		return false
	}
}
