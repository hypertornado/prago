package prago

import (
	"fmt"
)

// FieldType defines type of field
type fieldType struct {
	id string

	getViewFieldContent func(request *Request, val any) *viewFieldContent
	viewTemplate        string
	viewDataSource      func(*Request, *Field, any) any

	dbFieldDescription string

	allowedValues []string

	formHideLabel bool

	formTemplate      string
	formDataSource    func(*Field, UserData, string) any
	formValueStringer func(any) string

	listCellDataSource func(UserData, *Field, any) *listCell

	filterLayoutTemplate   string
	filterLayoutDataSource func(*Field, UserData) any

	naturalCellWidth int64
}

func (ft *fieldType) helpURL() string {
	if ft.id == "markdown" {
		return "/admin/help/markdown"
	}
	return ""
}

func (app *App) addFieldType(id string, fieldType *fieldType) {
	fieldType.id = id
	_, exist := app.fieldTypes[id]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", id))
	}

	if fieldType.getViewFieldContent == nil {
		if fieldType.viewTemplate == "" {
			panic(fmt.Sprintf("field type '%s' has empty viewTemplate", id))
		}

		if fieldType.viewDataSource == nil {
			panic(fmt.Sprintf("field type '%s' has empty viewDataSource", id))
		}
	}

	if fieldType.formTemplate == "" {
		panic(fmt.Sprintf("field type '%s' has empty formTemplate", id))
	}

	if fieldType.formValueStringer == nil {
		panic(fmt.Sprintf("field type '%s' has empty formValueStringer", id))
	}

	app.fieldTypes[id] = fieldType
}
func (f fieldType) isRelation() bool {
	if f.viewTemplate == "view_relation" {
		return true
	} else {
		return false
	}
}

type relationFormDataSource struct {
	App           *App
	RelatedID     string
	MultiRelation bool
}

func (rfds relationFormDataSource) Placeholder() string {
	relatedResource := rfds.App.getResourceByID(rfds.RelatedID)
	return fmt.Sprintf("Hledat v '%s'", relatedResource.pluralName("cs"))
}

func (app *App) initDefaultFieldTypes() {
	app.addFieldType("role", app.createRoleFieldType())

	app.addFieldType("string", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(defaultViewDataSource),

		formTemplate:      "form_input",
		formValueStringer: stringerString,
	})
	app.addFieldType("int64", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(numberViewDataSource),

		formTemplate:      "form_input_int",
		formValueStringer: stringerInt64,
	})
	app.addFieldType("float64", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(floatViewDataSource),

		formTemplate:      "form_input_float",
		formValueStringer: stringerFloat64,
	})
	app.addFieldType("bool", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(boolViewDataSource),

		formTemplate:      "form_input_checkbox",
		formValueStringer: stringerBool,

		formHideLabel: true,
	})

	app.addFieldType("text", &fieldType{
		viewTemplate:   "view_textarea",
		viewDataSource: stringerToDataSource(defaultViewDataSource),

		formTemplate:      "form_input_textarea",
		formValueStringer: stringerString,

		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(numberViewDataSource),

		formValueStringer: stringerInt64,

		formTemplate: "form_input_int",
	})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:       "view_cdn_file",
		viewDataSource:     cdnViewDataSource,
		formTemplate:       "form_input_cdnfile",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",
		formValueStringer:    stringerString,
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:      "view_image",
		viewDataSource:    fileViewDataSource,
		formTemplate:      "form_input_image",
		formDataSource:    imageFormDataSource(""),
		formValueStringer: stringerString,

		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		viewTemplate:       "view_image",
		viewDataSource:     fileViewDataSource,
		formTemplate:       "form_input_image",
		formDataSource:     imageFormDataSource(".jpg,.jpeg,.png"),
		formValueStringer:  stringerString,
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,

		naturalCellWidth: 60,
	})
	app.addFieldType("video", &fieldType{
		viewTemplate:   "view_video",
		viewDataSource: videoViewDataSource,

		formTemplate:      "form_input",
		formValueStringer: stringerString,

		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:      "view_markdown",
		viewDataSource:    markdownViewDataSource,
		formTemplate:      "form_input_markdown",
		formValueStringer: stringerString,

		listCellDataSource: markdownListDataSource,
	})
	app.addFieldType("place", &fieldType{
		viewTemplate:      "view_place",
		viewDataSource:    stringerToDataSource(defaultViewDataSource),
		formValueStringer: stringerString,

		formTemplate: "form_input_place",
	})

	app.addFieldType("relation", &fieldType{
		viewTemplate: "view_relation",
		viewDataSource: func(request *Request, f *Field, value any) any {
			valInt := value.(int64)
			return f.relationPreview(request, fmt.Sprintf("%d", valInt))
		},
		formTemplate: "form_input_relation",
		formDataSource: func(f *Field, userData UserData, value string) any {
			return relationFormDataSource{
				App:           app,
				RelatedID:     f.getRelatedID(),
				MultiRelation: false,
			}
		},
		formValueStringer: stringerInt64,
	})

	app.addFieldType("multirelation", &fieldType{
		viewTemplate: "view_relation",
		viewDataSource: func(request *Request, f *Field, value any) any {
			return f.relationPreview(request, value.(string))
		},
		formTemplate: "form_input_relation",
		formDataSource: func(f *Field, userData UserData, value string) any {
			return relationFormDataSource{
				App:           app,
				RelatedID:     f.getRelatedID(),
				MultiRelation: true,
			}
		},
		formValueStringer: stringerString,
	})

	app.addFieldType("Time", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(timestampViewDataSource),

		formTemplate:      "form_input_datetime",
		formValueStringer: stringerDate,
		naturalCellWidth:  130,
	})

	app.addFieldType("date", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(dateViewDataSource),

		formTemplate:      "form_input_date",
		formValueStringer: stringerDate,
		naturalCellWidth:  130,
	})

	app.addFieldType("timestamp", &fieldType{
		viewTemplate:   "view_text",
		viewDataSource: stringerToDataSource(timestampViewDataSource),

		formTemplate:      "form_input_timestamp",
		formValueStringer: stringerDateTime,
		naturalCellWidth:  130,
	})
}

func boolFilterLayoutDataSource(field *Field, userData UserData) any {
	return [][2]string{
		{"", ""},
		{"true", messages.Get(userData.Locale(), "yes")},
		{"false", messages.Get(userData.Locale(), "no")},
	}
}

func markdownViewDataSource(request *Request, f *Field, value any) any {
	return filterMarkdown(value.(string))
}
