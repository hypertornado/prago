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

func (app *App) getAdminNavigation(user User, code string) adminItemNavigation {
	tabs := []navigationTab{
		{
			Name:     messages.Get(user.Locale, "admin_signpost"),
			URL:      app.GetAdminURL(""),
			Selected: trueIfEqual(code, ""),
		},
	}

	for _, v := range app.rootActions {
		if v.method == "GET" {
			if app.Authorize(user, v.permission) {
				tabs = append(tabs, navigationTab{
					Name:     v.name(user.Locale),
					URL:      app.GetAdminURL(v.url),
					Selected: trueIfEqual(code, v.url),
				})
			}
		}
	}

	return adminItemNavigation{
		Tabs: tabs,
	}
}

func (resource Resource) getNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab
	for _, v := range resource.actions {
		if v.method == "GET" {
			if resource.App.Authorize(user, v.permission) {
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
			if resource.App.Authorize(user, v.permission) {
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

func (app *App) getSettingsNavigation(user User, code string) adminItemNavigation {
	var tabs []navigationTab

	tabs = append(tabs, navigationTab{
		Name:     messages.Get(user.Locale, "admin_settings"),
		URL:      app.GetAdminURL("settings"),
		Selected: trueIfEqual(code, "settings"),
	})

	tabs = append(tabs, navigationTab{
		Name:     messages.Get(user.Locale, "admin_password_change"),
		URL:      app.GetAdminURL("password"),
		Selected: trueIfEqual(code, "password"),
	})

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

/*
func createNavigationalItemHandler(action, templateName string, dataGenerator func(Resource, Request, User) interface{}) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		must(resource.App.Query().WhereIs("id", int64(id)).Get(item))

		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			App:          resource.App,
			Navigation:   resource.getItemNavigation(user, item, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}*/

//CreateNavigationalItemAction creates navigational item action
/*func createNavigationalItemAction(url string, name func(string) string, templateName string, dataGenerator func(Resource, Request, User) interface{}) Action {
	return Action{
		URL:     url,
		Name:    name,
		Handler: createNavigationalItemHandler(url, templateName, dataGenerator),
	}
}*/

/*
func createNavigationalHandler(action, templateName string, dataGenerator func(Resource, Request, User) interface{}) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(resource, request, user)
		}

		renderNavigationPage(request, adminNavigationPage{
			App:          resource.App,
			Navigation:   resource.getNavigation(user, action),
			PageTemplate: templateName,
			PageData:     data,
		})
	}
}

func createAdminHandler(action, templateName string, dataGenerator func(Request) interface{}, empty bool) func(Resource, Request, User) {
	return func(resource Resource, request Request, user User) {
		var data interface{}
		if dataGenerator != nil {
			data = dataGenerator(request)
		}

		adminNavigation := resource.App.getAdminNavigation(user, action)
		var name string
		for _, v := range adminNavigation.Tabs {
			if v.Selected {
				name = v.Name
			}
		}

		renderNavigationPage(request, adminNavigationPage{
			Name:         name,
			App:          resource.App,
			PageTemplate: templateName,
			PageData:     data,
			HideBox:      empty,
		})
	}
}*/
