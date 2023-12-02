package prago

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

func isTabVisible(tabs []tab, pos int) bool {
	if tabs[pos-1].Selected {
		return false
	}
	if tabs[pos].Selected {
		return false
	}
	return true
}

func (resourceData *resourceData) getResourceNavigation(userData UserData, code string) navigation {
	var tabs []tab
	for _, v := range resourceData.actions {
		if v.method == "GET" {
			if userData.Authorize(v.permission) {
				tabs = append(tabs, tab{
					Icon:     v.icon,
					Name:     v.name(userData.Locale()),
					URL:      resourceData.getURL(v.url),
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

func (resourceData *resourceData) getItemNavigation(userData UserData, item interface{}, code string) navigation {
	var tabs []tab
	for _, v := range resourceData.itemActions {
		if v.method == "GET" {
			name := v.name(userData.Locale())
			if v.url == "" {
				name = resourceData.previewer(userData, item).Name()
			}
			if userData.Authorize(v.permission) {
				tabs = append(tabs, tab{
					Icon:     v.icon,
					Name:     name,
					URL:      resourceData.getItemURL(item, v.url, userData),
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
		Icon:     "glyphicons-basic-431-log-in.svg",
		URL:      app.getAdminURL("user/login"),
		Selected: trueIfEqual(code, "login"),
	})

	tabs = append(tabs, tab{
		Name:     messages.Get(language, "admin_register"),
		Icon:     "glyphicons-basic-7-user-plus.svg",
		URL:      app.getAdminURL("user/registration"),
		Selected: trueIfEqual(code, "registration"),
	})

	tabs = append(tabs, tab{
		Name:     messages.Get(language, "admin_forgotten"),
		Icon:     "glyphicons-basic-45-key.svg",
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
