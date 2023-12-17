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
		strVal := field.fieldType.formStringer(ifaceVal)
		ret[field.id] = strVal
	}
	return ret
}

func (resource *Resource) addFormItems(item any, request *Request, form *Form) {
	editableValues := resource.getItemStringEditableValues(item, request)

fields:
	for _, field := range resource.fields {
		if !field.authorizeEdit(request) {
			continue fields
		}

		item := &FormItem{
			ID:       field.id,
			Icon:     field.getIcon(),
			Name:     field.name(request.Locale()),
			Template: field.fieldType.formTemplate,
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
			item.Data = field.fieldType.formDataSource(field, request)
		}

		if field.required {
			item.Required = true
		}

		form.AddItem(item)
	}
}
