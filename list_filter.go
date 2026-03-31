package prago

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
)

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

		if field.fieldType.filterLayoutTemplate == "filter_layout_select" {
			listFilterFormSelect(form, field, value, request)
		}

		if field.fieldType.isRelation() {
			listFilterFormRelation(form, field, value, request)
		}

		if field.filterLayout() == "filter_layout_date" {
			listFilterFormDate(form, field, value, request)
		}

		if field.filterLayout() == "filter_layout_boolean" {
			listFilterFormBoolean(form, field, value, request)
		}

		if field.filterLayout() == "filter_layout_text" {
			listFilterFormText(form, field, value, request)
		}

		if field.filterLayout() == "filter_layout_number" {
			listFilterFormNumber(form, field, value, request)
		}

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

		if field.fieldType.filterLayoutTemplate == "filter_layout_select" {
			listFilterFormSelectHandle(fv, field, request)
		}

		if field.fieldType.isRelation() {
			listFilterFormRelationHandle(fv, field, request)
		}

		if field.filterLayout() == "filter_layout_date" {
			listFilterFormDateHandle(fv, field, request)
		}

		if field.filterLayout() == "filter_layout_boolean" {
			listFilterFormBooleanHandle(fv, field, request)
		}

		if field.filterLayout() == "filter_layout_text" {
			listFilterFormTextHandle(fv, field, request)
		}

		if field.filterLayout() == "filter_layout_number" {
			listFilterFormNumberHandle(fv, field, request)
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
func listFilterFormRelation(form *Form, field *Field, value string, userData UserData) {
	items := form.AddRelationMultiple("items", field.name(userData.Locale()), field.getRelatedID())
	items.Focused = true
	items.Value = strings.ReplaceAll(value, ",", ";")
}

func listFilterFormDate(form *Form, field *Field, value string, userData UserData) {

	fieldName := field.name(userData.Locale())

	dates := strings.Split(value, ",")
	from := form.AddDatePicker("from", fieldName+" - Od")
	if len(dates) == 2 {
		from.Value = dates[0]
	}

	to := form.AddDatePicker("to", fieldName+" - Do")
	if len(dates) == 2 {
		to.Value = dates[1]
	}

}

func listFilterFormBoolean(form *Form, field *Field, value string, userData UserData) {
	var options [][2]string
	options = append(options, [2]string{"", ""})
	options = append(options, [2]string{"true", "✅ ano"})
	options = append(options, [2]string{"false", "ne"})
	form.AddRadio("value", field.name(userData.Locale()), options).Value = value
}

func listFilterFormText(form *Form, field *Field, value string, userData UserData) {
	item := form.AddTextInput("value", field.name(userData.Locale()))
	item.Focused = true
	item.Value = value
}

func listFilterFormNumber(form *Form, field *Field, value string, userData UserData) {
	item := form.AddNumberInput("value", field.name(userData.Locale()))
	item.Focused = true
	item.Value = value
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

func listFilterFormRelationHandle(fv FormValidation, field *Field, request *Request) {
	var values []string
	ints := MultirelationStringToArray(request.Param("items"))
	for _, i := range ints {
		values = append(values, fmt.Sprintf("%d", i))
	}
	fv.Data(listFilterGetResponse(strings.Join(values, ","), field, request))
}

func listFilterFormDateHandle(fv FormValidation, field *Field, request *Request) {
	if request.Param("from") != "" {
		_, err := time.Parse("2006-01-02", request.Param("from"))
		if err != nil {
			fv.AddItemError("from", "Neplatný formát")
		}
	}
	if request.Param("to") != "" {
		_, err := time.Parse("2006-01-02", request.Param("to"))
		if err != nil {
			fv.AddItemError("to", "Neplatný formát")
		}
	}

	if !fv.Valid() {
		return
	}

	val := fmt.Sprintf("%s,%s", request.Param("from"), request.Param("to"))
	if val == "," {
		val = ""
	}
	fv.Data(listFilterGetResponse(val, field, request))
}

func listFilterFormBooleanHandle(fv FormValidation, field *Field, request *Request) {
	fv.Data(listFilterGetResponse(request.Param("value"), field, request))
}

func listFilterFormTextHandle(fv FormValidation, field *Field, request *Request) {
	fv.Data(listFilterGetResponse(request.Param("value"), field, request))
}

func listFilterFormNumberHandle(fv FormValidation, field *Field, request *Request) {
	fv.Data(listFilterGetResponse(request.Param("value"), field, request))
}

func listFilterGetResponse(value string, field *Field, request *Request) (ret *ListFilterResponse) {
	ret = &ListFilterResponse{}
	ret.ID = value
	ret.Name = value

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

	if field.fieldType.isRelation() {
		var names []string
		if value != "" {
			ids := strings.Split(value, ",")
			for _, id := range ids {
				var name = fmt.Sprintf("#%s", id)
				item := field.relatedResource.query(context.Background()).ID(id)
				if item != nil {
					prev := field.relatedResource.previewer(request, item)
					name = prev.Name()
				}
				names = append(names, name)
			}
		}
		ret.Name = strings.Join(names, " nebo ")
	}

	if field.filterLayout() == "filter_layout_date" {
		var names []string
		if ret.ID != "" {
			fields := strings.Split(ret.ID, ",")
			if fields[0] != "" {
				t, err := time.Parse("2006-01-02", fields[0])
				if err == nil {
					names = append(names, fmt.Sprintf("od %s", t.Format("2. 1. 2006")))
				}
			}
			if fields[1] != "" {
				t, err := time.Parse("2006-01-02", fields[1])
				if err == nil {
					names = append(names, fmt.Sprintf("do %s", t.Format("2. 1. 2006")))
				}
			}
		}
		ret.Name = strings.Join(names, " · ")
	}

	if field.filterLayout() == "filter_layout_boolean" {
		if value == "true" {
			ret.Name = "✅ ano"
		}
		if value == "false" {
			ret.Name = "ne"
		}
	}

	return
}
