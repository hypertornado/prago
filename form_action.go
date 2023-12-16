package prago

import (
	"context"
	"fmt"
)

type formAction struct {
	validation    Validation
	formGenerator func(*Form, *Request)

	actionForm       *Action
	actionValidation *Action
}

func newFormAction(app *App, url string, injectForm func(*Form, *Request)) *formAction {
	ret := &formAction{
		actionForm:       newAction(app, url),
		actionValidation: newAction(app, url).Method("POST"),
	}

	ret.actionForm.childAction = ret.actionValidation

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

func (app *App) FormAction(url string, formGenerator func(*Form, *Request), validator Validation) *Action {
	fa := newFormAction(app, url, nil)

	fa.formGenerator = formGenerator
	fa.validation = validator

	app.rootActions = append(app.rootActions, fa.actionForm)
	app.rootActions = append(app.rootActions, fa.actionValidation)
	return fa.actionForm
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
			App:      app,
			Tabs:     app.getNologinNavigation(locale, id),
			FormData: form,
		})
	})

	app.accessController.routeHandler("POST", fmt.Sprintf("/admin/user/%s", id), func(request *Request) {
		requestValidator := newRequestValidation(request)
		validator(requestValidator)
		request.WriteJSON(200, requestValidator.validation)
	})

}

func ResourceFormAction[T any](app *App, url string, formGenerator func(*Form, *Request), validation Validation) *Action {
	resource := GetResource[T](app)
	return resource.data.FormAction(url, formGenerator, validation)
}

func (resourceData *resourceData) FormAction(url string, formGenerator func(*Form, *Request), validation Validation) *Action {
	action := newFormAction(resourceData.app, url, nil)

	action.actionForm.resourceData = resourceData
	action.actionValidation.resourceData = resourceData

	action.actionForm.Permission(resourceData.canView)
	action.actionValidation.Permission(resourceData.canView)

	action.formGenerator = formGenerator
	action.validation = validation

	resourceData.actions = append(resourceData.actions, action.actionForm)
	resourceData.actions = append(resourceData.actions, action.actionValidation)
	return action.actionForm
}

func ResourceFormItemAction[T any](
	app *App,
	url string,
	formGenerator func(*T, *Form, *Request),
	validation func(*T, ValidationContext),
) *Action {
	resourceData := GetResource[T](app).data
	return resourceData.formItemAction(
		url,
		func(a any, f *Form, r *Request) {
			formGenerator(a.(*T), f, r)
		},
		func(a any, vc ValidationContext) {
			validation(a.(*T), vc)
		},
	)
}

func (resourceData *resourceData) formItemAction(url string, formGenerator func(any, *Form, *Request), validation func(any, ValidationContext)) *Action {
	fa := newFormAction(resourceData.app, url, func(f *Form, r *Request) {
		item := resourceData.query(context.TODO()).ID(r.Param("id"))
		f.image = resourceData.previewer(r, item).ImageURL(r.r.Context())
	})

	fa.actionForm.resourceData = resourceData
	fa.actionValidation.resourceData = resourceData

	fa.actionForm.Permission(resourceData.canView)
	fa.actionValidation.Permission(resourceData.canView)

	fa.actionForm.isItemAction = true
	fa.actionValidation.isItemAction = true

	resourceData.itemActions = append(resourceData.itemActions, fa.actionForm)
	resourceData.itemActions = append(resourceData.itemActions, fa.actionValidation)

	fa.formGenerator = func(form *Form, request *Request) {
		item := resourceData.query(request.r.Context()).ID(request.Param("id"))
		formGenerator(item, form, request)
	}

	fa.validation = func(vc ValidationContext) {
		item := resourceData.query(vc.Context()).ID(vc.GetValue("id"))
		validation(item, vc)
	}

	return fa.actionForm
}
