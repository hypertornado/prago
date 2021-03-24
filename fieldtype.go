package prago

import (
	"fmt"
	"time"

	"github.com/hypertornado/prago/utils"
)

//FieldType defines type of field
type FieldType struct {
	ViewTemplate   string
	ViewDataSource func(*user, field, interface{}) interface{}

	DBFieldDescription string

	FormHideLabel bool

	FormTemplate   string
	FormDataSource func(field, *user) interface{}
	FormStringer   func(interface{}) string

	ListCellDataSource func(*user, field, interface{}) interface{}
	ListCellTemplate   string

	FilterLayoutTemplate   string
	FilterLayoutDataSource func(field, *user) interface{}
}

//AddFieldType adds field type
func (app *App) AddFieldType(name string, fieldType FieldType) {
	_, exist := app.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	app.fieldTypes[name] = fieldType
}

//IsRelation detects if field type is relation type
func (f FieldType) IsRelation() bool {
	if f.ViewTemplate == "admin_item_view_relation" {
		return true
	}
	return false
}

func (app *App) initDefaultFieldTypes() {
	app.AddFieldType("role", app.createRoleFieldType())

	app.AddFieldType("text", FieldType{
		ViewTemplate:       "admin_item_view_textarea",
		FormTemplate:       "admin_item_textarea",
		ListCellDataSource: textListDataSource,
	})
	app.AddFieldType("order", FieldType{})
	app.AddFieldType("date", FieldType{})

	app.AddFieldType("cdnfile", FieldType{
		ViewTemplate:   "admin_item_view_file",
		ViewDataSource: filesViewDataSource,
		FormTemplate:   "admin_file",
		//ListCellTemplate: "admin_item_view_file_cell",
		ListCellTemplate:   "admin_list_image",
		ListCellDataSource: defaultViewDataSource,

		FilterLayoutTemplate:   "filter_layout_select",
		FilterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.AddFieldType("file", FieldType{
		ViewTemplate:     "admin_item_view_image",
		FormTemplate:     "admin_item_image",
		FormDataSource:   createFilesEditDataSource(""),
		ListCellTemplate: "admin_list_image",

		FilterLayoutTemplate:   "filter_layout_select",
		FilterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.AddFieldType("image", FieldType{
		ViewTemplate:     "admin_item_view_image",
		FormTemplate:     "admin_item_image",
		FormDataSource:   createFilesEditDataSource(".jpg,.jpeg,.png"),
		ListCellTemplate: "admin_list_image",

		FilterLayoutTemplate:   "filter_layout_select",
		FilterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.AddFieldType("markdown", FieldType{
		ViewTemplate:       "admin_item_view_markdown",
		FormTemplate:       "admin_item_markdown",
		ListCellDataSource: markdownListDataSource,
		ListCellTemplate:   "admin_item_view_text",
	})
	app.AddFieldType("place", FieldType{
		ViewTemplate: "admin_item_view_place",
		FormTemplate: "admin_item_place",

		ListCellTemplate: "admin_item_view_text",
	})

	app.AddFieldType("relation", FieldType{
		ViewTemplate:     "admin_item_view_relation",
		ListCellTemplate: "admin_item_view_relation_cell",
		ViewDataSource:   getRelationViewData,

		FormTemplate: "admin_item_relation",
		FormDataSource: func(f field, u *user) interface{} {
			if f.Tags["prago-relation"] != "" {
				return columnName(f.Tags["prago-relation"])
			}
			return columnName(f.Name)
		},
	})

	app.AddFieldType("timestamp", FieldType{
		FormTemplate: "admin_item_timestamp",
		FormStringer: func(i interface{}) string {
			tm := i.(time.Time)
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02 15:04")
		},
	})
}

func boolFilterLayoutDataSource(field field, user *user) interface{} {
	return [][2]string{
		{"", ""},
		{"true", messages.Get(user.Locale, "yes")},
		{"false", messages.Get(user.Locale, "no")},
	}
}

func textListDataSource(user *user, f field, value interface{}) interface{} {
	return utils.Crop(value.(string), 100)
}

func createFilesEditDataSource(mimeTypes string) func(f field, u *user) interface{} {
	return func(f field, u *user) interface{} {
		return mimeTypes
	}
}

func markdownListDataSource(user *user, f field, value interface{}) interface{} {
	return utils.CropMarkdown(value.(string), 100)
}
