package prago

import (
	"reflect"
)

func (resource Resource) getDefaultOrder() (column string, desc bool) {
	for _, v := range resource.fieldArrays {
		add := false
		if v.ColumnName == "id" {
			add = true
		}
		if v.Tags["prago-type"] == "order" {
			add = true
		}
		if v.Tags["prago-order"] == "true" {
			add = true
		}
		if v.Tags["prago-order-desc"] == "true" {
			add = true
		}
		if v.Tags["prago-order"] == "false" {
			add = false
		}

		if add {
			column = v.ColumnName
			desc = false
			if v.Tags["prago-order-desc"] == "true" {
				desc = true
			}
		}
	}
	return
}

func (resource Resource) getItemStringEditableValues(item interface{}, user *user) map[string]string {
	itemVal := reflect.ValueOf(item).Elem()
	ret := make(map[string]string)
	for i, field := range resource.fieldArrays {
		if !field.authorizeEdit(user) {
			continue
		}
		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)
		strVal := field.fieldType.formStringer(ifaceVal)
		ret[field.ColumnName] = strVal
	}
	return ret
}

func (resource Resource) addFormItems(item interface{}, user *user, form *Form) {
	editableValues := resource.getItemStringEditableValues(item, user)

fields:
	for _, field := range resource.fieldArrays {
		if !field.authorizeEdit(user) {
			continue fields
		}

		item := &FormItem{
			ID:       field.ColumnName,
			Name:     field.HumanName(user.Locale),
			Template: field.fieldType.formTemplate,
		}
		if field.Description != nil {
			item.Description = field.Description(user.Locale)
		}
		item.AddUUID()

		if field.fieldType.formHideLabel {
			item.HiddenName = true
		}

		item.Value = editableValues[field.ColumnName]

		if field.fieldType.formDataSource != nil {
			item.Data = field.fieldType.formDataSource(*field, user)
		}

		form.AddItem(item)
	}
}
