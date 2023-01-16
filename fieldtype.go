package prago

import (
	"context"
	"fmt"
	"time"
)

// FieldType defines type of field
type fieldType struct {
	viewTemplate   string
	viewDataSource func(context.Context, *user, *Field, interface{}) interface{}

	dbFieldDescription string

	allowedValues []string

	formHideLabel  bool
	formTemplate   string
	formDataSource func(*Field, *user) interface{}
	formStringer   func(interface{}) string

	listCellDataSource func(*user, *Field, interface{}) listCell

	filterLayoutTemplate   string
	filterLayoutDataSource func(*Field, *user) interface{}

	fieldTypeIcon    string
	naturalCellWidth int64
}

func (app *App) addFieldType(name string, fieldType *fieldType) {
	_, exist := app.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	app.fieldTypes[name] = fieldType
}

// IsRelation detects if field type is relation type
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
		fieldTypeIcon:      "glyphicons-basic-101-text.svg",
	})
	app.addFieldType("order", &fieldType{})
	app.addFieldType("date", &fieldType{
		naturalCellWidth: 130,
	})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:       "admin_item_view_file",
		viewDataSource:     filesViewDataSource,
		formTemplate:       "admin_file",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",

		//filterLayoutTemplate:   "filter_layout_select",
		//filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:       "admin_item_view_image",
		formTemplate:       "admin_item_image",
		formDataSource:     createFilesEditDataSource(""),
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		viewTemplate:       "admin_item_view_image",
		formTemplate:       "admin_item_image",
		formDataSource:     createFilesEditDataSource(".jpg,.jpeg,.png"),
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,

		fieldTypeIcon:    "glyphicons-basic-38-picture.svg",
		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:       "admin_item_view_markdown",
		viewDataSource:     markdownViewDataSource,
		formTemplate:       "admin_item_markdown",
		listCellDataSource: markdownListDataSource,
		fieldTypeIcon:      "glyphicons-basic-692-font.svg",
		//listCellTemplate:   "admin_item_view_text",
	})
	app.addFieldType("place", &fieldType{
		viewTemplate:  "admin_item_view_place",
		formTemplate:  "admin_item_place",
		fieldTypeIcon: "glyphicons-basic-591-map-marker.svg",
		//listCellTemplate: "admin_item_view_text",
	})

	app.addFieldType("relation", &fieldType{
		viewTemplate: "admin_item_view_relation",
		//listCellTemplate: "admin_item_view_relation_cell",
		viewDataSource: func(ctx context.Context, user *user, f *Field, value interface{}) interface{} {
			return f.relationPreview(ctx, user, value.(int64))
		},
		//viewDataSource: getRelationViewData,
		formTemplate: "admin_item_relation",
		formDataSource: func(f *Field, u *user) interface{} {
			if f.tags["prago-relation"] != "" {
				return columnName(f.tags["prago-relation"])
			}
			return f.id
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
		naturalCellWidth: 130,
	})
}

func boolFilterLayoutDataSource(field *Field, user *user) interface{} {
	return [][2]string{
		{"", ""},
		{"true", messages.Get(user.Locale, "yes")},
		{"false", messages.Get(user.Locale, "no")},
	}
}

func createFilesEditDataSource(mimeTypes string) func(f *Field, u *user) interface{} {
	return func(f *Field, u *user) interface{} {
		return mimeTypes
	}
}

func markdownViewDataSource(ctx context.Context, user *user, f *Field, value interface{}) interface{} {
	return cropMarkdown(value.(string), 100)
}
