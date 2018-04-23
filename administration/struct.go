package administration

import (
	"errors"
	"go/ast"
	"reflect"
)

func (resource *Resource) newStructCache(item interface{}, fieldTypes map[string]FieldType) error {
	typ := reflect.TypeOf(item)
	if typ.Kind() != reflect.Struct {
		return errors.New("item is not a structure, but " + typ.Kind().String())
	}

	resource.fieldMap = make(map[string]*field)
	resource.fieldTypes = fieldTypes

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := newField(typ.Field(i), i, fieldTypes)
			if field.Tags["prago-type"] == "order" {
				resource.OrderFieldName = field.Name
				resource.OrderColumnName = field.ColumnName
			}
			resource.fieldArrays = append(resource.fieldArrays, field)
			resource.fieldMap[field.ColumnName] = field
		}
	}
	return nil
}

func (resource Resource) GetDefaultOrder() (column string, desc bool) {
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

func (resource Resource) GetForm(inValues interface{}, user User, filters ...fieldFilter) (*Form, error) {
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
			Name:      field.Name,
			NameHuman: field.Name,
			Template:  field.fieldType.FormTemplate,
		}
		item.AddUUID()

		if field.fieldType.FormHideLabel {
			item.HiddenName = true
		}
		item.Value = field.fieldType.FormStringer(ifaceVal)
		item.NameHuman = field.HumanName(user.Locale)

		if field.fieldType.FormDataSource != nil {
			item.Data = field.fieldType.FormDataSource(*field, user)
		}

		form.AddItem(item)
	}

	return form, nil
}
