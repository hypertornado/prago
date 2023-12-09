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
		actionForm:       newAction(app, url),
		actionValidation: newAction(app, url).Method("POST"),
	}

	ret.actionForm.ui(func(request *Request, pd *pageData) {
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
		pd.Form = form
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
		request.WriteJSON(200, rv.validation)
	})

	return ret
}

func (board *Board) FormAction(url string, formGenerator func(*Form, *Request), validator Validation) *FormAction {
	app := board.app
	fa := newFormAction(app, url, nil)

	fa.formGenerator = formGenerator
	fa.validation = validator

	fa.actionForm.parentBoard = board
	fa.actionValidation.parentBoard = board
	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa
}

func (app *App) nologinFormAction(id string, formHandler func(f *Form, r *Request), validator Validation) {
	app.accessController.routeHandler("GET", fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		if request.UserID() > 0 {
			request.Redirect("/admin")
			return
		}

		locale := localeFromRequest(request)
		form := NewForm("/admin/user/" + id)
		formHandler(form, request)

		renderPageNoLogin(request, &pageNoLogin{
			App:        app,
			Navigation: app.getNologinNavigation(locale, id),
			FormData:   form,
		})
	})

	app.accessController.routeHandler("POST", fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		requestValidator := newRequestValidation(request)
		validator(requestValidator)
		request.WriteJSON(200, requestValidator.validation)
	})

}

func ResourceFormAction[T any](app *App, url string, formGenerator func(*Form, *Request), validation Validation) *FormAction {
	resource := GetResource[T](app)
	return resource.data.FormAction(url, formGenerator, validation)
}

func (resourceData *resourceData) FormAction(url string, formGenerator func(*Form, *Request), validation Validation) *FormAction {
	action := newFormAction(resourceData.app, url, nil)

	action.actionForm.resourceData = resourceData
	action.actionValidation.resourceData = resourceData

	action.actionForm.Permission(resourceData.canView)
	action.actionValidation.Permission(resourceData.canView)

	action.formGenerator = formGenerator
	action.validation = validation

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

/*func (formAction *FormAction) Form(formGenerator func(*Form, *Request)) *FormAction {
	formAction.formGenerator = formGenerator
	return formAction
}

func (formAction *FormAction) Validation(validation Validation) *FormAction {
	formAction.validation = validation
	return formAction
}*/

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
