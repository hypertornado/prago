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

	listCellDataSource func(ud UserData, field *Field, item any) *listCell

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

	if fieldType.listCellDataSource == nil {
		panic(fmt.Sprintf("field type '%s' has empty listCellDataSource", id))
	}

	if fieldType.dbFieldDescription == "" {
		panic(fmt.Sprintf("field type '%s' has empty dbFieldDescription", id))
	}

	app.fieldTypes[id] = fieldType
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
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(defaultStringer),

		formTemplate:      "form_input",
		formValueStringer: stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),
	})
	app.addFieldType("int64", &fieldType{
		dbFieldDescription: "bigint(20)",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(numberStringer),

		formTemplate:      "form_input_int",
		formValueStringer: stringerInt64,

		listCellDataSource: basicCellDataSource(numberStringer),

		naturalCellWidth: 60,
	})
	app.addFieldType("float64", &fieldType{
		dbFieldDescription: "double",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(floatStringer),

		formTemplate:      "form_input_float",
		formValueStringer: stringerFloat64,

		listCellDataSource: basicCellDataSource(floatStringer),

		naturalCellWidth: 60,
	})
	app.addFieldType("bool", &fieldType{
		dbFieldDescription: "bool NOT NULL",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(boolStringer),

		formTemplate:      "form_input_checkbox",
		formValueStringer: stringerBool,

		listCellDataSource: basicCellDataSource(boolStringer),

		formHideLabel: true,

		naturalCellWidth: 60,
	})

	app.addFieldType("text", &fieldType{
		dbFieldDescription: "text",
		viewTemplate:       "view_textarea",
		viewDataSource:     stringerToDataSource(defaultStringer),

		formTemplate:      "form_input_textarea",
		formValueStringer: stringerString,

		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{
		dbFieldDescription: "bigint(20)",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(numberStringer),

		formValueStringer: stringerInt64,

		listCellDataSource: basicCellDataSource(numberStringer),

		formTemplate: "form_input_int",
	})

	app.addFieldType("cdnfile", &fieldType{
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_cdn_file",
		viewDataSource:     cdnViewDataSource,
		formTemplate:       "form_input_cdnfile",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",
		formValueStringer:    stringerString,
	})

	app.addFieldType("file", &fieldType{
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_image",
		viewDataSource:     fileViewDataSource,
		formTemplate:       "form_input_image",
		formDataSource:     imageFormDataSource(""),
		formValueStringer:  stringerString,

		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		dbFieldDescription: "text",
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
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_video",
		viewDataSource:     videoViewDataSource,

		formTemplate:      "form_input",
		formValueStringer: stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),

		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		dbFieldDescription: "text",
		viewTemplate:       "view_markdown",
		viewDataSource:     markdownViewDataSource,
		formTemplate:       "form_input_markdown",
		formValueStringer:  stringerString,

		listCellDataSource: markdownListDataSource,
	})
	app.addFieldType("place", &fieldType{
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_place",
		viewDataSource:     stringerToDataSource(defaultStringer),
		formValueStringer:  stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),

		formTemplate: "form_input_place",
	})

	app.addFieldType("relation", &fieldType{
		dbFieldDescription: "bigint(20)",
		viewTemplate:       "view_relation",
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

		listCellDataSource: relationCellViewData,

		naturalCellWidth: 150,
	})

	app.addFieldType("multirelation", &fieldType{
		dbFieldDescription: "varchar(255)",
		viewTemplate:       "view_relation",
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

		listCellDataSource: relationCellViewData,

		naturalCellWidth: 150,
	})

	app.addFieldType("date", &fieldType{
		dbFieldDescription: "date",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(dateStringer),

		formTemplate:      "form_input_date",
		formValueStringer: stringerDate,
		naturalCellWidth:  130,

		listCellDataSource: basicCellDataSource(dateStringer),
	})

	app.addFieldType("time", &fieldType{
		dbFieldDescription: "datetime",
		viewTemplate:       "view_text",
		viewDataSource:     stringerToDataSource(timeStringer),

		formTemplate:      "form_input_timestamp",
		formValueStringer: stringerDateTime,
		naturalCellWidth:  130,

		listCellDataSource: basicCellDataSource(timeStringer),
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
