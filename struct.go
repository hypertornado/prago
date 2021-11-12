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

func (resource Resource) getForm(inValues interface{}, request *Request, action string) (*Form, error) {
	user := request.user
	form := NewForm(action)
	itemVal := reflect.ValueOf(inValues).Elem()

fields:
	for i, field := range resource.fieldArrays {
		if !field.authorizeEdit(user) {
			continue fields
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

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
		item.Value = field.fieldType.formStringer(ifaceVal)

		if field.fieldType.formDataSource != nil {
			item.Data = field.fieldType.formDataSource(*field, user)
		}

		form.AddItem(item)
	}

	return form, nil
}
