package prago

import (
	"html/template"
)

type PageData struct {
	Language  string
	Version   string
	GoogleKey string

	LogoURL string

	CSSPaths        []string
	JavascriptPaths []string

	Icon  string
	Style string
	Name  string
	App   *App

	AllowSearch bool

	SearchQuery string

	Breadcrumbs *breadcrumbs

	ErrorMessage *pageMessage

	PageContent template.HTML

	MenuItems []*MenuItem

	FooterItems []template.HTML

	BoxHeader *BoxHeader

	Form *Form
	List *list

	Views []*View

	Pagination []*paginationItem

	BoardView *boardView

	NotificationsData string
	HTTPCode          int
	Hue               int64

	AnalyticsCode template.HTML
}

type pageMessage struct {
	Name string
}

func (pd *PageData) AddView(view *View) *PageData {
	pd.Views = append(pd.Views, view)
	return pd
}

func (pd *PageData) AddMenuItem(item *MenuItem) *PageData {
	pd.MenuItems = append(pd.MenuItems, item)
	return pd
}

func (app *App) HandlePage(url string, handler func(request *Request, pageData *PageData), constraints ...Constraint) {
	app.Handle("GET", url, func(request *Request) {
		pd := createPageData(request)
		handler(request, pd)
		pd.renderPage(request)
	}, constraints...)
}

func createPageData(request *Request) *PageData {
	page := &PageData{}
	page.App = request.app
	page.Language = request.Locale()
	page.GoogleKey = request.app.GoogleKey()
	page.Version = request.app.GetVersionString()

	for _, v := range request.app.cssPaths {
		page.CSSPaths = append(page.CSSPaths, v())
	}
	for _, v := range request.app.javascriptPaths {
		page.JavascriptPaths = append(page.JavascriptPaths, v())
	}

	page.NotificationsData = request.getNotificationsData()
	page.Hue = request.app.hue
	return page
}

func (page *PageData) renderPage(request *Request) {

	page.Icon, page.Style = page.GetIconAndStyle()
	if page.Icon == "" {
		page.Icon = request.app.icon
	}
	if page.Icon == "" {
		page.Icon = iconResource
	}

	page.Breadcrumbs = page.GetBreadcrumbs()

	title := page.GetTitle()
	if title != "" {
		page.Name = title
	}

	code := page.HTTPCode
	if code == 0 {
		code = 200
	}

	request.WriteHTML(code, request.app.adminTemplates, "layout", page)
}

func (page *PageData) GetIconColor() string {
	if page.Style == "" {
		return page.App.getBaseColor()
	}
	return getStyleColor(page.Style)
}
