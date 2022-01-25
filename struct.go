package prago

import (
	"reflect"
)

func (resource *Resource[T]) getDefaultOrder() (column string, desc bool) {
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

func (resource *Resource[T]) getItemStringEditableValues(item *T, user *user) map[string]string {
	itemVal := reflect.ValueOf(item).Elem()
	ret := make(map[string]string)
	for i, field := range resource.fields {
		if !field.authorizeEdit(user) && field.id != "id" {
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

func (resource *Resource[T]) addFormItems(item *T, user *user, form *Form) {
	editableValues := resource.getItemStringEditableValues(item, user)

fields:
	for _, field := range resource.fields {
		if !field.authorizeEdit(user) {
			continue fields
		}

		item := &FormItem{
			ID:       field.id,
			Name:     field.name(user.Locale),
			Template: field.fieldType.formTemplate,
		}
		if field.description != nil {
			item.Description = field.description(user.Locale)
		}
		item.AddUUID()

		if field.fieldType.formHideLabel {
			item.HiddenName = true
		}

		item.Value = editableValues[field.id]

		if field.fieldType.formDataSource != nil {
			item.Data = field.fieldType.formDataSource(field, user)
		}

		if field.required {
			item.Required = true
		}

		form.AddItem(item)
	}
}
