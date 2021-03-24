package prago

import (
	"fmt"
	"time"

	"github.com/hypertornado/prago/utils"
)

//FieldType defines type of field
type FieldType struct {
	viewTemplate   string
	viewDataSource func(*user, field, interface{}) interface{}

	dbFieldDescription string

	formHideLabel bool

	formTemplate   string
	formDataSource func(field, *user) interface{}
	formStringer   func(interface{}) string

	listCellDataSource func(*user, field, interface{}) interface{}
	listCellTemplate   string

	filterLayoutTemplate   string
	filterLayoutDataSource func(field, *user) interface{}
}

//FieldType adds field type
func (app *App) FieldType(name string) *FieldType {
	ft := &FieldType{}
	app.addFieldType(name, ft)
	return ft
}

func (ft *FieldType) AddViewTemplate(template string) *FieldType {
	ft.viewTemplate = template
	return ft
}

func (ft *FieldType) AddFormTemplate(template string) *FieldType {
	ft.formTemplate = template
	return ft
}

func (ft *FieldType) AddDBFieldDescription(description string) *FieldType {
	ft.dbFieldDescription = description
	return ft
}

func (app *App) addFieldType(name string, fieldType *FieldType) {
	_, exist := app.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	app.fieldTypes[name] = fieldType
}

//IsRelation detects if field type is relation type
func (f FieldType) IsRelation() bool {
	if f.viewTemplate == "admin_item_view_relation" {
		return true
	}
	return false
}

func (app *App) initDefaultFieldTypes() {
	app.addFieldType("role", app.createRoleFieldType())

	app.addFieldType("text", &FieldType{
		viewTemplate:       "admin_item_view_textarea",
		formTemplate:       "admin_item_textarea",
		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &FieldType{})
	app.addFieldType("date", &FieldType{})

	app.addFieldType("cdnfile", &FieldType{
		viewTemplate:   "admin_item_view_file",
		viewDataSource: filesViewDataSource,
		formTemplate:   "admin_file",
		//ListCellTemplate: "admin_item_view_file_cell",
		listCellTemplate:   "admin_list_image",
		listCellDataSource: defaultViewDataSource,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("file", &FieldType{
		viewTemplate:     "admin_item_view_image",
		formTemplate:     "admin_item_image",
		formDataSource:   createFilesEditDataSource(""),
		listCellTemplate: "admin_list_image",

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("image", &FieldType{
		viewTemplate:     "admin_item_view_image",
		formTemplate:     "admin_item_image",
		formDataSource:   createFilesEditDataSource(".jpg,.jpeg,.png"),
		listCellTemplate: "admin_list_image",

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("markdown", &FieldType{
		viewTemplate:       "admin_item_view_markdown",
		formTemplate:       "admin_item_markdown",
		listCellDataSource: markdownListDataSource,
		listCellTemplate:   "admin_item_view_text",
	})
	app.addFieldType("place", &FieldType{
		viewTemplate:     "admin_item_view_place",
		formTemplate:     "admin_item_place",
		listCellTemplate: "admin_item_view_text",
	})

	app.addFieldType("relation", &FieldType{
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

	app.addFieldType("timestamp", &FieldType{
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
