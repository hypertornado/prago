package prago

type page struct {
	Language     string
	Icon         string
	Name         string
	App          *App
	Navigation   navigation
	PageTemplate string
	PageData     interface{}
	HTTPCode     int
	Menu         menu
}

type pageNoLogin struct {
	App        *App
	Navigation navigation
	FormData   interface{}
}

func renderPage(request *Request, page page) {
	page.Language = request.Locale()
	page.Menu = request.app.getMenu(request, request.Request().URL.Path, request.csrfToken())

	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			page.Name = v.Name
			if v.Icon != "" {
				page.Icon = v.Icon
			}
		}
	}

	if page.Name == "" {
		page.Name = page.Menu.GetTitle()
	}

	if page.Icon == "" {
		page.Icon = request.app.icon
	}

	if page.Name == "" {
		page.Name = request.app.name(request.Locale())
	}

	request.SetData("page", page)

	code := page.HTTPCode
	if code == 0 {
		code = 200
	}

	request.RenderViewWithCode("layout", code)
}

func renderPageNoLogin(request *Request, page pageNoLogin) {
	var name string
	var icon string

	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			name = v.Name
			icon = v.Icon
		}
	}

	request.SetData("admin_title", name)
	request.SetData("admin_icon", icon)
	request.SetData("admin_page", page)
	request.RenderViewWithCode("layout_nologin", 200)
}
