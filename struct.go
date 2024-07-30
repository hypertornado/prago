package prago

import (
	"reflect"
)

func (resource *Resource) getDefaultOrder() (column string, desc bool) {
	for _, v := range resource.fields {
		add := false
		if v.id == "id" {
			add = true
		}
		if v.tags["prago-type"] == "order" {
			add = true
		}
		if v.tags["prago-order"] == "true" {
			add = true
		}
		if v.tags["prago-order-desc"] == "true" {
			add = true
		}
		if v.tags["prago-order"] == "false" {
			add = false
		}

		if add {
			column = v.id
			desc = false
			if v.tags["prago-order-desc"] == "true" {
				desc = true
			}
		}
	}
	return
}

func (resource *Resource) getItemStringEditableValues(item any, request *Request) map[string]string {
	itemVal := reflect.ValueOf(item).Elem()
	ret := make(map[string]string)
	for i, field := range resource.fields {
		if !field.authorizeEdit(request) && field.id != "id" {
			continue
		}
		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)
		strVal := field.fieldType.ft_formStringer(ifaceVal)
		ret[field.id] = strVal
	}
	return ret
}

func (form *Form) initWithResourceItem(resource *Resource, item any, request *Request) {
	editableValues := resource.getItemStringEditableValues(item, request)

	focusedField := request.Param("_focus")

	var firstField = true

	for _, field := range resource.fields {
		if !field.authorizeEdit(request) {
			continue
		}

		var focused = false
		if firstField && focusedField == "" {
			focused = true
		}
		firstField = false

		if field.id == focusedField {
			focused = true
		}

		item := &FormItem{
			ID:       field.id,
			Icon:     field.getIcon(),
			Name:     field.name(request.Locale()),
			Template: field.fieldType.formTemplate,
			Focused:  focused,
		}
		if field.description != nil {
			item.Description = field.description(request.Locale())
		}
		item.AddUUID()

		if field.fieldType.formHideLabel {
			item.HiddenName = true
		}

		item.Value = editableValues[field.id]

		if field.fieldType.formDataSource != nil {
			item.Data = field.fieldType.formDataSource(field, request, item.Value)
		}

		if field.required {
			item.Required = true
		}

		if field.formContentGenerator != nil {
			item.Content = field.formContentGenerator(item)
		}

		if field.fieldType.helpURL() != "" {
			item.HelpURL = field.fieldType.helpURL()
		}

		if field.helpURL != "" {
			item.HelpURL = field.helpURL
		}

		form.AddItem(item)
	}
}
