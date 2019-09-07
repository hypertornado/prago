package administration

import (
	"time"

	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
)

type FieldType struct {
	ViewTemplate   string
	ViewDataSource func(Resource, User, Field, interface{}) interface{}

	DBFieldDescription string

	FormHideLabel  bool
	FormTemplate   string
	FormDataSource func(Field, User) interface{}
	FormStringer   func(interface{}) string

	ListCellDataSource func(Resource, User, Field, interface{}) interface{}
	ListCellTemplate   string

	FilterLayoutTemplate   string
	FilterLayoutDataSource func(Field, User) interface{}
}

func (admin *Administration) addDefaultFieldTypes() {
	admin.AddFieldType("role", admin.createRoleFieldType())

	admin.AddFieldType("text", FieldType{
		ViewTemplate:       "admin_item_view_textarea",
		FormTemplate:       "admin_item_textarea",
		ListCellDataSource: textListDataSource,
	})
	admin.AddFieldType("order", FieldType{})
	admin.AddFieldType("date", FieldType{})

	admin.AddFieldType("file", FieldType{
		ViewTemplate:   "admin_item_view_file",
		ViewDataSource: filesViewDataSource,

		FormTemplate: "admin_file",

		ListCellTemplate: "admin_item_view_file_cell",

		FilterLayoutTemplate:   "filter_layout_select",
		FilterLayoutDataSource: boolFilterLayoutDataSource,
	})

	admin.AddFieldType("markdown", FieldType{
		ViewTemplate:       "admin_item_view_markdown",
		FormTemplate:       "admin_item_markdown",
		ListCellDataSource: markdownListDataSource,
		ListCellTemplate:   "admin_item_view_text",
	})
	admin.AddFieldType("image", FieldType{
		ViewTemplate:     "admin_item_view_image",
		FormTemplate:     "admin_item_image",
		ListCellTemplate: "admin_list_image",

		FilterLayoutTemplate:   "filter_layout_select",
		FilterLayoutDataSource: boolFilterLayoutDataSource,
	})
	admin.AddFieldType("place", FieldType{
		ViewTemplate: "admin_item_view_place",
		FormTemplate: "admin_item_place",

		ListCellTemplate: "admin_item_view_text",
	})

	admin.AddFieldType("relation", FieldType{
		ViewTemplate:     "admin_item_view_relation",
		ListCellTemplate: "admin_item_view_relation_cell",
		ViewDataSource:   getRelationViewData,

		FormTemplate: "admin_item_relation",
		FormDataSource: func(f Field, u User) interface{} {
			if f.Tags["prago-relation"] != "" {
				return columnName(f.Tags["prago-relation"])
			} else {
				return columnName(f.Name)
			}
		},
	})

	admin.AddFieldType("timestamp", FieldType{
		FormTemplate: "admin_item_timestamp",
		FormStringer: func(i interface{}) string {
			tm := i.(time.Time)
			if tm.IsZero() {
				return ""
			}
			/*if isZeroTime(tm) {
				return ""
			}*/
			return tm.Format("2006-01-02 15:04")
		},
	})
}

func boolFilterLayoutDataSource(field Field, user User) interface{} {
	return [][2]string{
		{"", ""},
		{"true", messages.Messages.Get(user.Locale, "yes")},
		{"false", messages.Messages.Get(user.Locale, "no")},
	}
}

func textListDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return utils.Crop(value.(string), 100)
}

func markdownListDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return utils.CropMarkdown(value.(string), 100)
}
