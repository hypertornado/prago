package prago

import "strings"

func listFilterFormAdd(form *Form, field *Field, value string, request *Request, autofocus bool) {

	filterLayout := field.fieldType.filterLayoutTemplate
	if filterLayout == "filter_layout_select" {
		listFilterFormSelect(form, field, value, request, autofocus)
	}

	if filterLayout == "filter_layout_relation" {
		listFilterFormRelation(form, field, value, request, autofocus)
	}

	if filterLayout == "filter_layout_date" {
		listFilterFormDate(form, field, value, request, autofocus)
	}

	if filterLayout == "filter_layout_boolean" {
		listFilterFormBoolean(form, field, value, request, autofocus)
	}

	if filterLayout == "filter_layout_text" {
		listFilterFormText(form, field, value, request, autofocus)
	}

	if filterLayout == "filter_layout_number" {
		listFilterFormNumber(form, field, value, request, autofocus)
	}

}

func listFilterFormSelect(form *Form, field *Field, value string, userData UserData, autofocus bool) {
	valMap := make(map[string]bool)
	for _, item := range strings.Split(value, ",") {
		valMap[item] = true
	}

	var dataItems = field.fieldType.filterLayoutDataSource(field, userData).([]*FormOption)
	for _, item := range dataItems {
		if item.ID == "" {
			continue
		}
		checkbox := form.AddCheckbox(field.id+"_"+item.ID, field.GetName("cs")+" - "+item.Name)
		checkbox.Icon = item.Icon
		checkbox.Style = item.Style
		checkbox.Color = item.Color
		if valMap[item.ID] {
			checkbox.SetValue("on")
		}
	}
}
func listFilterFormRelation(form *Form, field *Field, value string, userData UserData, autofocus bool) {
	item := form.
		AddRelationMultiple(field.id, field.name(userData.Locale()), field.getRelatedID()).
		SetValue(strings.ReplaceAll(value, ",", ";"))

	if autofocus {
		item.SetFocused()
	}
}

func listFilterFormDate(form *Form, field *Field, value string, userData UserData, _ bool) {

	fieldName := field.name(userData.Locale())

	dates := strings.Split(value, ",")
	from := form.AddDatePicker(field.id+"_from", fieldName+" - Od")
	if len(dates) == 2 {
		from.SetValue(dates[0])
	}

	to := form.AddDatePicker(field.id+"_to", fieldName+" - Do")
	if len(dates) == 2 {
		to.SetValue(dates[1])
	}

}

func listFilterFormBoolean(form *Form, field *Field, value string, userData UserData, _ bool) {

	var options []*FormOption = []*FormOption{
		{},
		{
			ID:    "true",
			Name:  messages.Get(userData.Locale(), "yes_plain"),
			Icon:  iconCheckbox,
			Style: "create",
		},
		{
			ID:    "false",
			Name:  messages.Get(userData.Locale(), "no_plain"),
			Style: "destroy",
		},
	}

	form.AddRadioOptions(field.id, field.name(userData.Locale()), options).SetValue(value)
}

func listFilterFormText(form *Form, field *Field, value string, userData UserData, autofocus bool) {
	item := form.
		AddText(field.id, field.name(userData.Locale())).
		SetValue(value)

	if autofocus {
		item.SetFocused()
	}

}

func listFilterFormNumber(form *Form, field *Field, value string, userData UserData, autofocus bool) {
	item := form.
		AddNumber(field.id, field.name(userData.Locale())).
		SetValue(value)

	if autofocus {
		item.SetFocused()
	}
}
