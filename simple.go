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
	Icon              string

	Tabs []*Button

	Sections []*SimpleSection

	FormData *Form

	PrimaryButton *Button

	FooterText template.HTML
}

type SimpleSection struct {
	Name        string
	Description string
}

const defaultBackgroundImageURL = "https://images.unsplash.com/photo-1519677100203-a0e668c92439?q=80&w=3540&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"

func renderPageSimple(request *Request, page *pageDataSimple) {
	var name string = page.Name
	var icon string

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
	request.WriteHTML(200, request.app.adminTemplates, "simple", page)
}
