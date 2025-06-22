package prago

import (
	"html/template"
)

type pageData struct {
	Language  string
	Version   string
	GoogleKey string

	CSSPaths        []string
	JavascriptPaths []string

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

	NotificationsData string
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
	page.GoogleKey = request.app.GoogleKey()
	page.Version = request.app.GetVersionString()

	for _, v := range request.app.cssPaths {
		page.CSSPaths = append(page.CSSPaths, v())
	}
	for _, v := range request.app.javascriptPaths {
		page.JavascriptPaths = append(page.JavascriptPaths, v())
	}

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

	request.WriteHTML(code, request.app.adminTemplates, "layout", page)
}
