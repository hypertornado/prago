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

func (resource Resource) getForm(inValues interface{}, user *user) (*form, error) {
	form := newForm()
	form.Method = "POST"
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

		item := &formItem{
			Name:      field.ColumnName,
			NameHuman: field.HumanName(user.Locale),
			Template:  field.fieldType.formTemplate,
		}
		item.AddUUID()

		if field.fieldType.formHideLabel {
			item.HiddenName = true
		}
		item.Value = field.fieldType.formStringer(ifaceVal)
		//item.NameHuman = field.HumanName(user.Locale)

		if field.fieldType.formDataSource != nil {
			item.Data = field.fieldType.formDataSource(*field, user)
		}

		form.AddItem(item)
	}

	return form, nil
}
