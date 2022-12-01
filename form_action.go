package prago

import "fmt"

type FormAction struct {
	validation    Validation
	formGenerator func(*Form, *Request)

	actionForm       *Action
	actionValidation *Action
}

func newFormAction(app *App, url string, injectForm func(*Form, *Request)) *FormAction {
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
		form.action = ret.actionForm

		if injectForm != nil {
			injectForm(form, request)
		}

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
	fa := newFormAction(app, url, nil)
	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa
}

func (app *App) nologinFormAction(id string, formHandler func(f *Form, r *Request), validator Validation) {
	app.accessController.get(fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		locale := localeFromRequest(request)
		form := NewForm("/admin/user/" + id)
		formHandler(form, request)

		renderPage(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, id),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

	app.accessController.post(fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		requestValidator := newRequestValidation(request)
		validator(requestValidator)
		request.RenderJSON(requestValidator.validation)
	})

}

func (resource *Resource[T]) FormAction(url string) *FormAction {
	return resource.data.FormAction(url)
}

func (resourceData *resourceData) FormAction(url string) *FormAction {
	action := newFormAction(resourceData.app, url, nil)

	action.actionForm.resourceData = resourceData
	action.actionValidation.resourceData = resourceData

	action.actionForm.Permission(resourceData.canView)
	action.actionValidation.Permission(resourceData.canView)

	resourceData.actions = append(resourceData.actions, action.actionForm)
	resourceData.actions = append(resourceData.actions, action.actionValidation)
	return action
}

func (formAction *FormAction) Name(name func(string) string) *FormAction {
	formAction.actionForm.Name(name)
	return formAction
}

func (formAction *FormAction) Icon(icon string) *FormAction {
	formAction.actionForm.icon = icon
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

func (formAction *FormAction) userMenu() *FormAction {
	formAction.actionForm.userMenu()
	return formAction
}

func (formAction *FormAction) priority() *FormAction {
	formAction.actionForm.priority()
	return formAction
}
