package prago

type FormAction struct {
	validation    Validation
	formGenerator func(*Form, *Request)

	actionForm       *Action
	actionValidation *Action
}

func newFormAction(app *App, url string) *FormAction {
	ret := &FormAction{
		actionForm:       newAction(app, url).Template("admin_form"),
		actionValidation: newAction(app, url).Method("POST"),
	}

	ret.actionForm.DataSource(func(request *Request) interface{} {
		if ret.formGenerator == nil {
			panic("No form set for this FormAction")
		}

		form := NewForm(request.r.URL.Path)
		form.AddCSRFToken(request)
		ret.formGenerator(form, request)
		return form
	})

	ret.actionValidation.Handler(func(request *Request) {
		if ret.validation == nil {
			panic("No validation set for this FormAction")
		}

		rv := newRequestValidation(request)
		if request.csrfToken() != rv.GetValue("_csrfToken") {
			panic("wrong csrf token")
		}
		ret.validation(rv)
		request.RenderJSON(rv.validation)
	})

	return ret
}

func (app *App) FormAction(url string) *FormAction {
	fa := newFormAction(app, url)
	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa
}

func (resource *Resource) FormAction(url string) *FormAction {
	fa := newFormAction(resource.app, url)

	fa.actionForm.resource = resource
	fa.actionValidation.resource = resource

	fa.actionForm.Permission(resource.canView)
	fa.actionValidation.Permission(resource.canView)

	resource.actions = append(resource.actions, fa.actionForm)
	resource.actions = append(resource.actions, fa.actionValidation)

	return fa
}

func (formAction *FormAction) Name(name func(string) string) *FormAction {
	formAction.actionForm.Name(name)
	return formAction
}

func (formAction *FormAction) Form(formGenerator func(*Form, *Request)) *FormAction {
	formAction.formGenerator = formGenerator
	return formAction
}

func (formAction *FormAction) Validation(validation Validation) *FormAction {
	formAction.validation = validation
	return formAction
}

func (formAction *FormAction) Permission(permission Permission) *FormAction {
	formAction.actionForm.Permission(permission)
	formAction.actionValidation.Permission(permission)
	return formAction
}

func (formAction *FormAction) IsWide() *FormAction {
	formAction.actionForm.IsWide()
	return formAction
}

func (formAction *FormAction) userMenu() *FormAction {
	formAction.actionForm.userMenu()
	return formAction
}
