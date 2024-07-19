package prago

import (
	"fmt"
	"time"
)

// FieldType defines type of field
type fieldType struct {
	viewTemplate   string
	viewDataSource func(*Request, *Field, interface{}) interface{}

	dbFieldDescription string

	allowedValues []string

	formHideLabel  bool
	formTemplate   string
	formDataSource func(*Field, UserData) interface{}
	formStringer   func(interface{}) string

	listCellDataSource func(UserData, *Field, interface{}) listCell

	filterLayoutTemplate   string
	filterLayoutDataSource func(*Field, UserData) interface{}

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

func (f fieldType) IsRelation() bool {
	if f.viewTemplate == "view_relation" {
		return true
	} else {
		return false
	}
}

type relationFormDataSource struct {
	RelatedID     string
	MultiRelation bool
}

func (app *App) initDefaultFieldTypes() {
	app.addFieldType("role", app.createRoleFieldType())

	app.addFieldType("text", &fieldType{
		viewTemplate:       "view_textarea",
		formTemplate:       "form_input_textarea",
		listCellDataSource: textListDataSource,
		fieldTypeIcon:      "glyphicons-basic-101-text.svg",
	})
	app.addFieldType("order", &fieldType{})
	app.addFieldType("date", &fieldType{
		naturalCellWidth: 130,
	})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:       "view_file",
		viewDataSource:     filesViewDataSource,
		formTemplate:       "form_input_cdnfile",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:       "view_image",
		formTemplate:       "form_input_image",
		formDataSource:     createFilesEditDataSource(""),
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		viewTemplate:       "view_image",
		formTemplate:       "form_input_image",
		formDataSource:     createFilesEditDataSource(".jpg,.jpeg,.png"),
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,

		fieldTypeIcon:    "glyphicons-basic-38-picture.svg",
		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:       "view_markdown",
		viewDataSource:     markdownViewDataSource,
		formTemplate:       "form_input_markdown",
		listCellDataSource: markdownListDataSource,
		fieldTypeIcon:      "glyphicons-basic-692-font.svg",
	})
	app.addFieldType("place", &fieldType{
		viewTemplate:  "view_place",
		formTemplate:  "form_input_place",
		fieldTypeIcon: "glyphicons-basic-591-map-marker.svg",
	})

	app.addFieldType("relation", &fieldType{
		viewTemplate: "view_relation",
		viewDataSource: func(request *Request, f *Field, value interface{}) interface{} {
			valInt := value.(int64)
			return f.relationPreview(request, fmt.Sprintf("%d", valInt))
		},
		formTemplate: "form_input_relation",
		formDataSource: func(f *Field, userData UserData) interface{} {
			return relationFormDataSource{
				RelatedID:     f.getRelatedID(),
				MultiRelation: false,
			}
		},
	})

	app.addFieldType("multirelation", &fieldType{
		viewTemplate: "view_relation",
		viewDataSource: func(request *Request, f *Field, value interface{}) interface{} {
			return f.relationPreview(request, value.(string))
		},
		formTemplate: "form_input_relation",
		formDataSource: func(f *Field, userData UserData) interface{} {
			return relationFormDataSource{
				RelatedID:     f.getRelatedID(),
				MultiRelation: true,
			}
		},
	})

	app.addFieldType("timestamp", &fieldType{
		formTemplate: "form_input_timestamp",
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

func boolFilterLayoutDataSource(field *Field, userData UserData) interface{} {
	return [][2]string{
		{"", ""},
		{"true", messages.Get(userData.Locale(), "yes")},
		{"false", messages.Get(userData.Locale(), "no")},
	}
}

func createFilesEditDataSource(mimeTypes string) func(*Field, UserData) interface{} {
	return func(f *Field, userData UserData) interface{} {
		return mimeTypes
	}
}

func markdownViewDataSource(request *Request, f *Field, value interface{}) interface{} {
	return filterMarkdown(value.(string))
}
