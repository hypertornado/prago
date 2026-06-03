package prago

import (
	"html/template"
)

type PageDataSimple struct {
	Request *Request

	BackButtons []*Button

	PreName  string
	Name     string
	PostName string

	BannerImageID string

	Sections []*SimpleSection

	Description template.HTML
	Text        template.HTML

	Form func(*Form)

	PrimaryButton *Button

	AnalyticsCode template.HTML

	FooterText template.HTML
}

type SimpleHandler struct {
	URL string

	Handler        func(pds *PageDataSimple)
	FormValidation func(fv FormValidation, request *Request)
}

func (app *App) HandleSimple(handler *SimpleHandler) {

	if handler.FormValidation != nil {
		app.router.route("POST", handler.URL, app.appController, func(request *Request) {

			rv := newFormValidation()
			if request.csrfToken() != request.Param("_csrfToken") {
				panic("wrong csrf token")
			}
			handler.FormValidation(rv, request)
			request.WriteJSON(200, rv.validationData)
		})
	}

	app.router.route("GET", handler.URL, app.appController, func(request *Request) {

		pd := &PageDataSimple{
			Request: request,
		}
		handler.Handler(pd)

		var form *Form
		if pd.Form != nil {
			form = app.NewForm(request.r.URL.Path)
			form.AddCSRFToken(request)
			pd.Form(form)
		}

		locale := localeFromRequest(request)

		var bannerImageURL string
		if pd.BannerImageID != "" {
			bannerImageURL = app.largeImage(pd.BannerImageID)
		}

		renderPageSimple(request, &pageDataSimple{
			Language: locale,
			App:      app,

			BackButtons: pd.BackButtons,

			PreName:     pd.PreName,
			Name:        pd.Name,
			PostName:    pd.PostName,
			Description: pd.Description,

			BannerImageURL: bannerImageURL,

			Sections: pd.Sections,

			Text:     pd.Text,
			FormData: form,

			PrimaryButton: pd.PrimaryButton,

			AnalyticsCode: pd.AnalyticsCode,

			FooterText: pd.FooterText,
		})
	})
}

func (pds *PageDataSimple) BackButton(btn *Button) {
	pds.BackButtons = append(pds.BackButtons, btn)
}
