package prago

import (
	"fmt"
)

type ListFilterResponse struct {
	ID   string
	Name string
}

func (app *App) initListFilter() {

	app.PopupForm("_list-fiter-item", func(form *Form, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		field := resource.fieldMap[request.Param("field")]
		if !request.Authorize(field.canView) {
			panic("not allowed")
		}

		value := request.Param("value")
		form.AddHidden("_resource").SetValue(resource.id)
		form.AddHidden("_field").SetValue(field.id)

		listFilterFormAdd(form, field, value, request, false)

		form.AddSubmit(fmt.Sprintf("Filtrovat pole „%s“", field.name(request.Locale())))

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("_resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		field := resource.fieldMap[request.Param("_field")]
		if !request.Authorize(field.canView) {
			panic("not allowed")
		}

		listFilterFormHandle(fv, field, request)

	}).Permission(loggedPermission).Name(Unlocalized("Filtrovat"))
}
