package prago

import (
	"fmt"
)

type Enum struct {
	ID                string
	Name              func(string) string
	DescriptionBefore func(string) string
	DescriptionAfter  func(string) string
	Icon              string
	Color             string
	Style             string
}

func (enum *Enum) GetName(locale string) string {
	if enum.Name != nil {
		return enum.Name(locale)
	}
	return enum.ID
}

func (app *App) AddEnumShort(name string, items [][2]string) {
	var arr []*Enum
	for _, v := range items {
		var itemName = v[1]
		arr = append(arr, &Enum{
			ID: v[0],
			Name: func(string) string {
				return itemName
			},
		})
	}
	app.AddEnum(name, arr)
}

func enumsToFormOptions(enums []*Enum, userData UserData) (ret []*FormOption) {
	for _, enum := range enums {
		option := &FormOption{
			ID:   enum.ID,
			Name: enum.ID,

			Icon:  enum.Icon,
			Color: enum.Color,
			Style: enum.Style,
		}
		if enum.Name != nil {
			option.Name = enum.Name(userData.Locale())
		}
		if enum.DescriptionBefore != nil {
			option.DescriptionBefore = enum.DescriptionBefore(userData.Locale())
		}
		if enum.DescriptionAfter != nil {
			option.DescriptionAfter = enum.DescriptionAfter(userData.Locale())
		}

		ret = append(ret, option)
	}
	return ret
}

func (app *App) AddEnum(name string, items []*Enum) {

	var allowedValues []string
	for _, v := range items {
		allowedValues = append(allowedValues, v.ID)
	}

	app.addFieldType(name, &fieldType{
		getViewFieldContent: func(request *Request, val any) *viewFieldContent {
			strVal := val.(string)
			ret := &viewFieldContent{
				Name: strVal,
			}
			for _, v := range items {
				if strVal == v.ID {
					ret.Name = v.GetName(request.Locale())
					return ret
				}
			}
			return ret

		},

		viewDataSource: func(request *Request, f *Field, value any) any {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return v.GetName(request.Locale())
				}
			}
			return value
		},

		allowedValues: allowedValues,

		formTemplate: "form_input_select",
		formDataSource: func(f *Field, userData UserData, value string) any {
			return enumsToFormOptions(items, userData)
		},
		formValueStringer: stringerString,

		listCellDataSource: func(userData UserData, f *Field, value any) *listCell {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return &listCell{Name: v.GetName(userData.Locale())}
				}
			}

			return &listCell{Name: fmt.Sprintf("%v", value)}
		},

		filterLayoutTemplate: "filter_layout_select",
		filterLayoutDataSource: func(f *Field, userData UserData) any {
			var ret [][2]string
			if len(items) == 0 || items[0].ID != "" {
				ret = append(ret, [2]string{"", ""})
			}
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.GetName(userData.Locale()),
				})
			}
			return ret
		},
	})
}
