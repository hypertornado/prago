package administration

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"time"
)

type structCache struct {
	typ             reflect.Type
	fieldArrays     []*field
	fieldMap        map[string]*field
	fieldTypes      map[string]FieldType
	OrderFieldName  string
	OrderColumnName string
}

func (resource *Resource) newStructCache(item interface{}, fieldTypes map[string]FieldType) error {
	typ := reflect.TypeOf(item)
	if typ.Kind() != reflect.Struct {
		return errors.New("item is not a structure, but " + typ.Kind().String())
	}

	resource.fieldMap = make(map[string]*field)
	resource.fieldTypes = fieldTypes

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := newField(typ.Field(i), i)
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

func (resource Resource) GetForm(inValues interface{}, user User, visible structFieldFilter, editable structFieldFilter) (*Form, error) {
	form := NewForm()

	form.Method = "POST"
	itemVal := reflect.ValueOf(inValues).Elem()

	for i, field := range resource.fieldArrays {
		if !visible(resource, user, *field) {
			continue
		}
		if !editable(resource, user, *field) {
			continue
		}

		var ifaceVal interface{}

		item := &FormItem{
			Name:        field.Name,
			NameHuman:   field.Name,
			SubTemplate: "admin_item_input",
		}

		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

		t, found := resource.fieldTypes[field.Tags["prago-type"]]
		if found {
			item.SubTemplate = t.FormSubTemplate
			if t.ValuesSource != nil {
				item.Values = (*t.ValuesSource)()
			}

			switch ifaceVal.(type) {
			case string:
				item.Value = ifaceVal.(string)
			case int64:
				item.Value = fmt.Sprintf("%d", ifaceVal.(int64))
			default:
				panic("unknown typ")
			}

		} else {
			switch field.Typ.Kind() {
			case reflect.Struct:
				if field.Typ == reflect.TypeOf(time.Now()) {
					tm := ifaceVal.(time.Time)
					if field.Tags["prago-type"] == "timestamp" {
						item.SubTemplate = "admin_item_timestamp"
						item.Value = tm.Format("2006-01-02 15:04")
					} else {
						item.SubTemplate = "admin_item_date"
						item.Value = tm.Format("2006-01-02")
					}
				}
			case reflect.Bool:
				item.SubTemplate = "admin_item_checkbox"
				if ifaceVal.(bool) {
					item.Value = "on"
				}
				item.HiddenName = true
			case reflect.String:
				item.Value = ifaceVal.(string)
				switch field.Tags["prago-type"] {
				case "text":
					item.SubTemplate = "admin_item_textarea"
				case "markdown":
					item.Template = "admin_item_markdown"
				case "image":
					item.SubTemplate = "admin_item_image"
				case "place":
					item.SubTemplate = "admin_item_place"
				}
			case reflect.Int64:
				item.Value = fmt.Sprintf("%d", ifaceVal.(int64))
				switch field.Tags["prago-type"] {
				case "relation":
					item.SubTemplate = "admin_item_relation"
					if field.Tags["prago-relation"] != "" {
						item.Values = columnName(field.Tags["prago-relation"])
					} else {
						item.Values = columnName(item.Name)
					}
				}
			case reflect.Float64:
				item.Value = fmt.Sprintf("%f", ifaceVal.(float64))
			default:
				panic("Wrong type" + field.Typ.Kind().String())
			}
		}

		item.NameHuman = field.HumanName(user.Locale)

		form.AddItem(item)
	}

	return form, nil
}
