package prago

type pageData struct {
	Language string
	Version  string

	Icon       string
	Name       string
	App        *App
	Navigation navigation

	Messages []pageMessage

	PageTemplate string
	PageData     interface{}
	Menu         menu

	Form  *Form
	List  *list
	Views []view

	BoardView *BoardView

	HelpIcons []string

	NotificationsData string
	JavaScripts       []string
	HTTPCode          int
}

type pageMessage struct {
	Name string
}

func createPageData(request *Request) *pageData {
	page := &pageData{}
	page.App = request.app
	page.Language = request.Locale()
	page.Version = request.app.version
	page.Menu = request.app.getMenu(request, request.Request().URL.Path, request.csrfToken())

	page.JavaScripts = request.app.javascripts
	page.NotificationsData = request.getNotificationsData()
	return page
}

func (page *pageData) renderPage(request *Request) {

	for _, v := range page.Navigation.Tabs {
		if v.Selected {
			page.Name = v.Name
			if v.Icon != "" {
				page.Icon = v.Icon
			}
		}
	}

	title := page.Menu.GetTitle()
	if title != "" {
		page.Name = title
	}

	if page.Icon == "" {
		page.Icon = request.app.icon
	}

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
