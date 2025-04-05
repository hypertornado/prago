package prago

import (
	"context"
	"fmt"
)

type formAction struct {
	formGenerator  func(*Form, *Request)
	formValidation func(FormValidation, *Request)

	actionForm       *Action
	actionValidation *Action
}

func newFormAction(app *App, url string, injectForm func(*Form, *Request)) *formAction {
	ret := &formAction{
		actionForm:       newAction(app, url),
		actionValidation: newAction(app, url).Method("POST"),
	}

	ret.actionForm.icon = iconForm
	ret.actionForm.isFormAction = true

	ret.actionForm.childAction = ret.actionValidation

	ret.actionForm.ui(func(request *Request, pd *pageData) {
		if ret.formGenerator == nil {
			panic("No form set for this FormAction")
		}

		form := app.NewForm(request.r.URL.Path)
		form.AddCSRFToken(request)
		form.action = ret.actionForm

		if injectForm != nil {
			injectForm(form, request)
		}

		ret.formGenerator(form, request)
		pd.Form = form
	})

	ret.actionValidation.addHandler(func(request *Request) {
		if ret.formValidation == nil {
			panic("No validation set for this FormAction")
		}

		rv := newFormValidation()
		if request.csrfToken() != request.Param("_csrfToken") {
			panic("wrong csrf token")
		}
		ret.formValidation(rv, request)
		request.WriteJSON(200, rv.validationData)
	})

	return ret
}

func ActionForm(app *App, url string, formGenerator func(*Form, *Request), validator func(FormValidation, *Request)) *Action {
	fa := newFormAction(app, url, nil)

	fa.formGenerator = formGenerator
	fa.formValidation = validator

	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa.actionForm
}

func (app *App) nologinFormAction(path string, formHandler func(f *Form, r *Request), validator func(FormValidation, *Request)) {
	app.accessController.routeHandler("GET", fmt.Sprintf("/admin/user/%s", path), func(request *Request) {
		if request.UserID() > 0 {
			request.Redirect("/admin")
			return
		}

		locale := localeFromRequest(request)
		form := app.NewForm("/admin/user/" + path)
		formHandler(form, request)

		renderPageNoLogin(request, &pageNoLogin{
			App:      app,
			Tabs:     app.getNologinNavigation(locale, path),
			FormData: form,
		})
	})

	app.accessController.routeHandler("POST", fmt.Sprintf("/admin/user/%s", path), func(request *Request) {
		requestValidator := newFormValidation()
		validator(requestValidator, request)
		request.WriteJSON(200, requestValidator.validationData)
	})

}

func ActionResourceForm[T any](app *App, url string, formGenerator func(*Form, *Request), validation func(FormValidation, *Request)) *Action {
	resource := getResource[T](app)
	return resource.formAction(url, formGenerator, validation)
}

func (resource *Resource) formAction(url string, formGenerator func(*Form, *Request), validation func(FormValidation, *Request)) *Action {
	action := newFormAction(resource.app, url, nil)

	action.actionForm.resource = resource
	action.actionValidation.resource = resource

	action.actionForm.Permission(resource.canView)
	action.actionValidation.Permission(resource.canView)

	action.formGenerator = formGenerator
	action.formValidation = validation

	resource.actions = append(resource.actions, action.actionForm)
	resource.actions = append(resource.actions, action.actionValidation)
	return action.actionForm
}

func ActionResourceItemForm[T any](
	app *App,
	url string,
	formGenerator func(*T, *Form, *Request),
	validation func(*T, FormValidation, *Request),
) *Action {
	resource := getResource[T](app)
	return resource.formItemAction(
		url,
		func(a any, f *Form, r *Request) {
			formGenerator(a.(*T), f, r)
		},
		func(a any, vc FormValidation, request *Request) {
			validation(a.(*T), vc, request)
		},
	)
}

func (resource *Resource) formItemAction(url string, formGenerator func(any, *Form, *Request), validation func(any, FormValidation, *Request)) *Action {
	fa := newFormAction(resource.app, url, func(f *Form, r *Request) {
		item := resource.query(context.TODO()).ID(r.Param("id"))
		f.image = resource.previewer(r, item).ImageURL(r.r.Context())
	})

	fa.actionForm.resource = resource
	fa.actionValidation.resource = resource

	fa.actionForm.Permission(resource.canView)
	fa.actionValidation.Permission(resource.canView)

	fa.actionForm.isItemAction = true
	fa.actionValidation.isItemAction = true

	resource.itemActions = append(resource.itemActions, fa.actionForm)
	resource.itemActions = append(resource.itemActions, fa.actionValidation)

	fa.formGenerator = func(form *Form, request *Request) {
		item := resource.query(request.r.Context()).ID(request.Param("id"))
		formGenerator(item, form, request)
	}

	fa.formValidation = func(vc FormValidation, request *Request) {
		item := resource.query(request.Request().Context()).ID(request.Param("id"))
		validation(item, vc, request)
	}

	return fa.actionForm
}

func PopupForm(app *App, url string, formGenerator func(*Form, *Request), validation func(FormValidation, *Request)) *Action {
	fa := newFormAction(app, url, nil)

	fa.actionForm.parentBoard = nil

	fa.formGenerator = formGenerator
	fa.formValidation = validation

	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)

	return fa.actionForm
}
