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

		if add == true {
			column = v.ColumnName
			desc = false
			if v.Tags["prago-order-desc"] == "true" {
				desc = true
			}
		}
	}
	return
}

func (resource Resource) getForm(inValues interface{}, user User, filters ...fieldFilter) (*Form, error) {
	filters = append(filters, defaultVisibilityFilter)
	filters = append(filters, defaultEditabilityFilter)
	form := NewForm()
	form.Method = "POST"
	itemVal := reflect.ValueOf(inValues).Elem()

fields:
	for i, field := range resource.fieldArrays {
		for _, filter := range filters {
			if !filter(resource, user, *field) {
				continue fields
			}
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

		item := &FormItem{
			Name:      field.ColumnName,
			NameHuman: field.HumanName(user.Locale),
			Template:  field.fieldType.FormTemplate,
		}
		item.AddUUID()

		if field.fieldType.FormHideLabel {
			item.HiddenName = true
		}
		item.Value = field.fieldType.FormStringer(ifaceVal)
		//item.NameHuman = field.HumanName(user.Locale)

		if field.fieldType.FormDataSource != nil {
			item.Data = field.fieldType.FormDataSource(*field, user)
		}

		form.AddItem(item)
	}

	return form, nil
}
