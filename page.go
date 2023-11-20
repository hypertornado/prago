package prago

type page struct {
	Language string
	Version  string

	Icon         string
	Name         string
	App          *App
	Navigation   navigation
	PageTemplate string
	PageData     interface{}
	Menu         menu

	NotificationsData string
	JavaScripts       []string
	HTTPCode          int
}

func renderPage(request *Request, page page) {
	page.Language = request.Locale()
	page.Version = request.app.version
	page.Menu = request.app.getMenu(request, request.Request().URL.Path, request.csrfToken())

	//fmt.Println(page.Menu)

	/*data, err := json.MarshalIndent(page.Menu, " ", " ")
	must(err)
	fmt.Println(string(data))*/

	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			page.Name = v.Name
			if v.Icon != "" {
				page.Icon = v.Icon
			}
		}
	}

	title := page.Menu.GetTitle()
	if page.Name == "" {
		page.Name = title
	} else {
		if page.Name != title {
			page.Name += " – " + title
		}
	}

	if page.Icon == "" {
		page.Icon = request.app.icon
	}

	if page.Name == "" {
		page.Name = request.app.name(request.Locale())
	}

	page.JavaScripts = request.app.javascripts
	page.NotificationsData = request.getNotificationsData()

	code := page.HTTPCode
	if code == 0 {
		code = 200
	}

	request.WriteHTML(code, "prago_layout", page)
}

type pageNoLogin struct {
	Language string
	Version  string
	App      *App

	NotificationsData string
	Title             string
	Icon              string

	Navigation navigation
	FormData   interface{}
}

func renderPageNoLogin(request *Request, page *pageNoLogin) {
	var name string
	var icon string

	page.Language = localeFromRequest(request)
	page.Version = request.app.version

	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			name = v.Name
			icon = v.Icon
		}
	}

	page.NotificationsData = request.getNotificationsData()
	page.Title = name
	page.Icon = icon
	request.WriteHTML(200, "layout_nologin", page)
}
