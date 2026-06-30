package prago

func PopupForm(app *App, url string, formGenerator func(form *Form, request *Request), validation func(fv FormValidation, request *Request)) *Action {
	fa := newFormAction(app, url, nil)
	fa.actionForm.parentBoard = nil
	fa.formGenerator = formGenerator
	fa.formValidation = validation
	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa.actionForm
}
