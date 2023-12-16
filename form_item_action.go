package prago

import "context"

func FormItemAction[T any](
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
