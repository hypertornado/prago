package prago

import (
	"fmt"
	"html/template"
	"math/rand"
)

type pageData struct {
	Language  string
	Version   string
	GoogleKey string

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
	if request.app.DevelopmentMode() {
		page.Version += fmt.Sprintf("-development-%d", rand.Intn(10000000000))
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

type pageNoLogin struct {
	CodeName string
	Language string
	Version  string
	App      *App

	BackgroundImageURL string

	NotificationsData string
	Title             string
	Icon              string

	Tabs     []*tab
	FormData interface{}
}

const defaultBackgroundImageURL = "https://images.unsplash.com/photo-1519677100203-a0e668c92439?q=80&w=3540&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"

func renderPageNoLogin(request *Request, page *pageNoLogin) {
	var name string
	var icon string

	page.Language = localeFromRequest(request)
	page.Version = request.app.version

	var err error
	page.BackgroundImageURL, err = request.app.getSetting("background_image_url")
	must(err)

	if page.BackgroundImageURL == "" {
		page.BackgroundImageURL = defaultBackgroundImageURL
	}

	for _, v := range page.Tabs {
		if v.Selected {
			name = v.Name
			icon = v.Icon
		}
	}

	page.CodeName = request.app.codeName

	page.NotificationsData = request.getNotificationsData()
	page.Title = name
	page.Icon = icon
	request.WriteHTML(200, request.app.adminTemplates, "layout_nologin", page)
}
