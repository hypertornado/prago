package prago

import "fmt"

type FormAction struct {
	validation    Validation
	formGenerator func(*Form, *Request)

	actionForm       *Action
	actionValidation *Action
}

func newFormAction(app *App, url string) *FormAction {
	ret := &FormAction{
		actionForm:       newAction(app, url).Template("admin_form").IsWide(),
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
			HideBox:      true,
		})
	})

	app.accessController.post(fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		requestValidator := newRequestValidation(request)
		validator(requestValidator)
		request.RenderJSON(requestValidator.validation)
	})

}

func (resource *Resource[T]) FormAction(url string) *FormAction {
	return newResourceFormAction(resource.Resource, url)
}

func newResourceFormAction(resource *resource, url string) *FormAction {
	fa := newFormAction(resource.app, url)

	fa.actionForm.resource = resource
	fa.actionValidation.resource = resource

	fa.actionForm.Permission(resource.canView)
	fa.actionValidation.Permission(resource.canView)

	resource.actions = append(resource.actions, fa.actionForm)
	resource.actions = append(resource.actions, fa.actionValidation)
	return fa
}

func (resource *Resource[T]) FormItemAction(url string) *FormAction {
	return newResourceItemFormAction(resource.Resource, url)
}

func newResourceItemFormAction(resource *resource, url string) *FormAction {
	fa := newFormAction(resource.app, url)

	fa.actionForm.resource = resource
	fa.actionValidation.resource = resource

	fa.actionForm.Permission(resource.canView)
	fa.actionValidation.Permission(resource.canView)

	fa.actionForm.isItemAction = true
	fa.actionValidation.isItemAction = true

	resource.itemActions = append(resource.itemActions, fa.actionForm)
	resource.itemActions = append(resource.itemActions, fa.actionValidation)
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

func (formAction *FormAction) priority() *FormAction {
	formAction.actionForm.priority()
	return formAction
}
