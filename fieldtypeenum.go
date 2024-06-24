package prago

import (
	"fmt"
)

type FieldTypeEnum struct {
	ID   string
	Name func(string) string
}

func (app *App) AddEnumFieldType(name string, items [][2]string) {
	var arr []FieldTypeEnum
	for _, v := range items {
		var itemName = v[1]
		arr = append(arr, FieldTypeEnum{
			ID: v[0],
			Name: func(string) string {
				return itemName
			},
		})
	}
	app.AddEnumFieldTypeLocalized(name, arr)
}

func (app *App) AddEnumFieldTypeLocalized(name string, items []FieldTypeEnum) {

	var allowedValues []string
	for _, v := range items {
		allowedValues = append(allowedValues, v.ID)
	}

	app.addFieldType(name, &fieldType{
		viewDataSource: func(request *Request, f *Field, value interface{}) interface{} {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return v.Name(request.Locale())
				}
			}
			return value
		},

		allowedValues: allowedValues,

		formTemplate: "admin_item_select",
		formDataSource: func(f *Field, userData UserData) interface{} {
			var ret [][2]string
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.Name(userData.Locale()),
				})
			}
			return ret
		},

		listCellDataSource: func(userData UserData, f *Field, value interface{}) listCell {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return listCell{Name: v.Name(userData.Locale())}
				}
			}

			return listCell{Name: fmt.Sprintf("%v", value)}
		},

		filterLayoutTemplate: "filter_layout_select",
		filterLayoutDataSource: func(f *Field, userData UserData) interface{} {
			var ret [][2]string
			if len(items) == 0 || items[0].ID != "" {
				ret = append(ret, [2]string{"", ""})
			}
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.Name(userData.Locale()),
				})
			}
			return ret
		},
		fieldTypeIcon: "glyphicons-basic-299-circle-selected.svg",
	})
}
