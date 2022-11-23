package prago

import "fmt"

// FieldTypeEnum enum type of field
type FieldTypeEnum struct {
	ID   string
	Name func(string) string
}

// AddEnumFieldType adds enum field type
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

// AddEnumFieldTypeLocalized adds localized enum field
func (app *App) AddEnumFieldTypeLocalized(name string, items []FieldTypeEnum) {

	var allowedValues []string
	for _, v := range items {
		allowedValues = append(allowedValues, v.ID)
	}

	app.addFieldType(name, &fieldType{
		viewDataSource: func(user *user, f *Field, value interface{}) interface{} {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return v.Name(user.Locale)
				}
			}

			return value
		},

		allowedValues: allowedValues,

		formTemplate: "admin_item_select",
		formDataSource: func(f *Field, u *user) interface{} {
			var ret [][2]string
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.Name(u.Locale),
				})
			}
			return ret
		},

		cellDataSource: func(user *user, f *Field, value interface{}) cellViewData {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return cellViewData{Name: v.Name(user.Locale)}
				}
			}

			return cellViewData{Name: fmt.Sprintf("%v", value)}
		},

		filterLayoutTemplate: "filter_layout_select",
		filterLayoutDataSource: func(f *Field, user *user) interface{} {
			var ret [][2]string
			if len(items) == 0 || items[0].ID != "" {
				ret = append(ret, [2]string{"", ""})
			}
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.Name(user.Locale),
				})
			}
			return ret
		},
	})
}
