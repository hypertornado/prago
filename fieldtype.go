package prago

import (
	"fmt"
	"time"
)

//FieldType defines type of field
type fieldType struct {
	viewTemplate   string
	viewDataSource func(*user, field, interface{}) interface{}

	dbFieldDescription string

	allowedValues []string

	formHideLabel bool

	formTemplate   string
	formDataSource func(field, *user) interface{}
	formStringer   func(interface{}) string

	listCellDataSource func(*user, field, interface{}) interface{}
	listCellTemplate   string

	filterLayoutTemplate   string
	filterLayoutDataSource func(field, *user) interface{}
}

func (app *App) addFieldType(name string, fieldType *fieldType) {
	_, exist := app.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	app.fieldTypes[name] = fieldType
}

//IsRelation detects if field type is relation type
func (f fieldType) IsRelation() bool {
	if f.viewTemplate == "admin_item_view_relation" {
		return true
	} else {
		return false
	}
}

func (app *App) initDefaultFieldTypes() {
	app.addFieldType("role", app.createRoleFieldType())

	app.addFieldType("text", &fieldType{
		viewTemplate:       "admin_item_view_textarea",
		formTemplate:       "admin_item_textarea",
		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{})
	app.addFieldType("date", &fieldType{})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:   "admin_item_view_file",
		viewDataSource: filesViewDataSource,
		formTemplate:   "admin_file",
		//ListCellTemplate: "admin_item_view_file_cell",
		listCellTemplate:   "admin_list_image",
		listCellDataSource: defaultViewDataSource,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:     "admin_item_view_image",
		formTemplate:     "admin_item_image",
		formDataSource:   createFilesEditDataSource(""),
		listCellTemplate: "admin_list_image",

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("image", &fieldType{
		viewTemplate:     "admin_item_view_image",
		formTemplate:     "admin_item_image",
		formDataSource:   createFilesEditDataSource(".jpg,.jpeg,.png"),
		listCellTemplate: "admin_list_image",

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:       "admin_item_view_markdown",
		formTemplate:       "admin_item_markdown",
		listCellDataSource: markdownListDataSource,
		listCellTemplate:   "admin_item_view_text",
	})
	app.addFieldType("place", &fieldType{
		viewTemplate:     "admin_item_view_place",
		formTemplate:     "admin_item_place",
		listCellTemplate: "admin_item_view_text",
	})

	app.addFieldType("relation", &fieldType{
		viewTemplate:     "admin_item_view_relation",
		listCellTemplate: "admin_item_view_relation_cell",
		viewDataSource:   getRelationViewData,
		formTemplate:     "admin_item_relation",
		formDataSource: func(f field, u *user) interface{} {
			if f.Tags["prago-relation"] != "" {
				return columnName(f.Tags["prago-relation"])
			}
			return columnName(f.Name)
		},
	})

	app.addFieldType("timestamp", &fieldType{
		formTemplate: "admin_item_timestamp",
		formStringer: func(i interface{}) string {
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
	return crop(value.(string), 100)
}

func createFilesEditDataSource(mimeTypes string) func(f field, u *user) interface{} {
	return func(f field, u *user) interface{} {
		return mimeTypes
	}
}

func markdownListDataSource(user *user, f field, value interface{}) interface{} {
	return cropMarkdown(value.(string), 100)
}
