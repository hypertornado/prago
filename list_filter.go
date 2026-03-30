package prago

import "strings"

type ListFilterResponse struct {
	ID   string
	Name string
}

func (app *App) initListFilter() {

	PopupForm(app, "_list-fiter-item", func(form *Form, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		field := resource.fieldMap[request.Param("field")]
		if !request.Authorize(field.canView) {
			panic("not allowed")
		}

		value := request.Param("value")
		form.AddHidden("_resource").Value = resource.id
		form.AddHidden("_field").Value = field.id
		//form.AddHidden("_value").Value = value

		if field.fieldType.filterLayoutTemplate == "filter_layout_select" {
			listFilterFormSelect(form, field, value, request)
		}

		form.AddSubmit("Filtrovat")

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("_resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		field := resource.fieldMap[request.Param("_field")]
		if !request.Authorize(field.canView) {
			panic("not allowed")
		}

		if field.fieldType.filterLayoutTemplate == "filter_layout_select" {
			listFilterFormSelectHandle(fv, field, request)
		}

	}).Permission(loggedPermission).Name(unlocalized("Filtrovat"))
}

func listFilterFormSelect(form *Form, field *Field, value string, userData UserData) {
	valMap := make(map[string]bool)
	for _, item := range strings.Split(value, ",") {
		valMap[item] = true
	}

	var dataItems = field.fieldType.filterLayoutDataSource(field, userData).([][2]string)
	for _, item := range dataItems {
		if item[0] == "" {
			continue
		}
		checkbox := form.AddCheckbox(item[0], item[1])
		if valMap[item[0]] {
			checkbox.Value = "on"
		}
	}
}

func listFilterFormSelectHandle(fv FormValidation, field *Field, request *Request) {
	var values []string
	var dataItems = field.fieldType.filterLayoutDataSource(field, request).([][2]string)
	for _, item := range dataItems {
		if item[0] == "" {
			continue
		}
		if request.Param(item[0]) == "on" {
			values = append(values, item[0])
		}
	}

	fv.Data(listFilterGetResponse(strings.Join(values, ","), field, request))
}

func listFilterGetResponse(value string, field *Field, request *Request) (ret *ListFilterResponse) {
	ret = &ListFilterResponse{}
	ret.ID = value

	if field.fieldType.filterLayoutTemplate == "filter_layout_select" {

		var names []string

		nameMap := map[string]string{}

		var dataItems = field.fieldType.filterLayoutDataSource(field, request).([][2]string)
		for _, item := range dataItems {
			nameMap[item[0]] = item[1]

		}

		valItems := strings.Split(value, ",")
		for _, valItem := range valItems {
			names = append(names, nameMap[valItem])
		}
		ret.Name = strings.Join(names, " nebo ")
	}

	return
}
