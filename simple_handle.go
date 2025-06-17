package prago

import "html/template"

type PageDataSimple struct {
	Request *Request

	BackButton *Button

	PreName  string
	Name     string
	PostName string

	Sections []*SimpleSection

	Description template.HTML
	Text        template.HTML

	Form           func(*Form)
	FormValidation func(FormValidation, *Request)

	PrimaryButton *Button

	FooterText template.HTML
}

type SimpleHandler struct {
	URL string

	Handler        func(*PageDataSimple)
	FormValidation func(FormValidation, *Request)
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
			form = app.NewForm(handler.URL)
			form.AddCSRFToken(request)
			pd.Form(form)
		}

		locale := localeFromRequest(request)

		renderPageSimple(request, &pageDataSimple{
			Language: locale,
			App:      app,

			BackButton: pd.BackButton,

			PreName:     pd.PreName,
			Name:        pd.Name,
			PostName:    pd.PostName,
			Description: pd.Description,

			Sections: pd.Sections,

			Text:     pd.Text,
			FormData: form,

			PrimaryButton: pd.PrimaryButton,

			FooterText: pd.FooterText,
		})
	})
}
