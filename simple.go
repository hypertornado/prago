package prago

import (
	"html/template"
)

type pageDataSimple struct {
	CodeName string
	Language string
	Version  string
	App      *App

	CSSPaths        []string
	JavascriptPaths []string

	BackgroundImageURL string

	BackButton *Button

	PreName     string
	Name        string
	PostName    string
	Description template.HTML
	Text        template.HTML

	NotificationsData string
	Title             string
	Logo              string
	//Icon              string

	Tabs []*Button

	Sections []*SimpleSection

	FormData *Form

	PrimaryButton *Button

	AnalyticsCode template.HTML

	FooterText template.HTML
}

type SimpleSection struct {
	Icon        string
	Name        string
	Description string
	Text        template.HTML
	LogoURL     string

	Table *Table

	PrimaryButton *Button
	Buttons       []*Button
}

const defaultBackgroundImageURL = "https://images.unsplash.com/photo-1519677100203-a0e668c92439?q=80&w=3540&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"

func renderPageSimple(request *Request, page *pageDataSimple) {
	var name string = page.Name

	page.Language = localeFromRequest(request)
	page.Version = request.app.GetVersionString()

	for _, v := range request.app.cssPaths {
		page.CSSPaths = append(page.CSSPaths, v())
	}
	for _, v := range request.app.javascriptPaths {
		page.JavascriptPaths = append(page.JavascriptPaths, v())
	}

	var err error
	page.BackgroundImageURL, err = request.app.getSetting("background_image_url")
	must(err)
	if page.BackgroundImageURL == "" {
		page.BackgroundImageURL = defaultBackgroundImageURL
	}

	page.Logo, err = request.app.getSetting("icon_image_url")
	must(err)

	for _, v := range page.Tabs {
		if v.Selected {
			name = v.Name
		}
	}

	page.CodeName = request.app.codeName

	page.NotificationsData = request.getNotificationsData()
	page.Title = name
	request.WriteHTML(200, request.app.adminTemplates, "simple", page)
}
