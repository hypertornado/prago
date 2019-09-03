package administration

type FieldTypeEnum struct {
	ID   string
	Name func(string) string
}

//func(string) string

func (admin *Administration) AddEnumFieldType(name string, items [][2]string) {
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
	admin.AddEnumFieldTypeLocalized(name, arr)
}

func (admin *Administration) AddEnumFieldTypeLocalized(name string, items []FieldTypeEnum) {
	admin.AddFieldType(name, FieldType{
		ViewDataSource: func(resource Resource, user User, f Field, value interface{}) interface{} {
			str := value.(string)
			for _, v := range items {
				if str == v.ID {
					return v.Name(user.Locale)
				}
			}

			return value
		},

		FormTemplate: "admin_item_select",
		FormDataSource: func(f Field, u User) interface{} {
			var ret [][2]string
			for _, v := range items {
				ret = append(ret, [2]string{
					v.ID,
					v.Name(u.Locale),
				})
			}
			return ret
		},

		FilterLayoutTemplate: "filter_layout_select",
		FilterLayoutDataSource: func(f Field, user User) interface{} {
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
