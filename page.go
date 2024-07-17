package prago

import (
	"fmt"
	"html/template"
	"math/rand"
)

type pageData struct {
	Language string
	Version  string

	Icon string
	Name string
	App  *App

	SearchQuery string

	Breadcrumbs *breadcrumbs

	Messages []pageMessage

	PageContent template.HTML

	Menu *menu

	Form  *Form
	List  *list
	Views []*view

	BoardView *boardView

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
	if request.app.DevelopmentMode() {
		page.Version += fmt.Sprintf("-development-%d", rand.Intn(10000000000))
	}

	page.JavaScripts = request.app.javascripts
	page.NotificationsData = request.getNotificationsData()
	return page
}

func (page *pageData) renderPage(request *Request) {
	if page.Menu == nil {
		page.Menu = request.app.getMenu(request, nil)
	}

	page.Icon = page.Menu.GetIcon()
	if page.Icon == "" {
		page.Icon = request.app.icon
	}

	page.Breadcrumbs = page.Menu.GetBreadcrumbs()

	title := page.Menu.GetTitle()
	if title != "" {
		page.Name = title
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

	Tabs     []*tab
	FormData interface{}
}

func renderPageNoLogin(request *Request, page *pageNoLogin) {
	var name string
	var icon string

	page.Language = localeFromRequest(request)
	page.Version = request.app.version

	for _, v := range page.Tabs {
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
