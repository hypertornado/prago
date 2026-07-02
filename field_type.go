package prago

import (
	"fmt"
	"html/template"

	"github.com/golang-commonmark/markdown"
)

type fieldType struct {
	id string

	getViewFieldContent func(request *Request, field *Field, value any) *viewFieldContent

	dbFieldDescription string

	allowedValues []string

	formHideLabel bool

	formTemplate      string
	formDataSource    func(*Field, UserData, string) any
	formValueStringer func(value any) string

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
		panic(fmt.Sprintf("field type '%s' has empty getViewFieldContent", id))
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

		getViewFieldContent: stringerToViewFieldContent(defaultStringer),

		formTemplate:      "form_input",
		formValueStringer: stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),
	})
	app.addFieldType("int64", &fieldType{
		dbFieldDescription: "bigint(20)",

		getViewFieldContent: func(request *Request, field *Field, val any) *viewFieldContent {
			intVal := val.(int64)
			if intVal == 0 {
				return nil
			}
			return &viewFieldContent{
				Name: humanizeNumber(intVal),
			}
		},

		formTemplate:      "form_input_int",
		formValueStringer: stringerInt64,

		listCellDataSource: basicCellDataSource(numberStringer),

		naturalCellWidth: 60,
	})
	app.addFieldType("float64", &fieldType{
		dbFieldDescription: "double",

		getViewFieldContent: func(request *Request, field *Field, val any) *viewFieldContent {
			floatVal := val.(float64)
			if floatVal == 0 {
				return nil
			}
			return &viewFieldContent{
				Name: humanizeFloat(floatVal, request.Locale()),
			}
		},

		formTemplate:      "form_input_float",
		formValueStringer: stringerFloat64,

		listCellDataSource: basicCellDataSource(floatStringer),

		naturalCellWidth: 60,
	})
	app.addFieldType("bool", &fieldType{
		dbFieldDescription: "bool NOT NULL",
		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			if !value.(bool) {
				return nil
			}

			return &viewFieldContent{
				Name:  messages.Get(request.Locale(), "yes_plain"),
				Icon:  "glyphicons-basic-153-square-checkbox.svg",
				Style: "create",
			}
		},

		formTemplate: "form_input_checkbox",
		formValueStringer: func(value any) string {
			if value.(bool) {
				return "on"
			}
			return ""

		},

		listCellDataSource: func(ud UserData, field *Field, item any) *listCell {
			ret := &listCell{}

			if item.(bool) {
				ret.Style = "create"
				ret.Icon = "glyphicons-basic-153-square-checkbox.svg"
			}

			return ret

		},

		formHideLabel: true,

		naturalCellWidth: 20,
	})

	app.addFieldType("text", &fieldType{
		dbFieldDescription:  "text",
		getViewFieldContent: stringerToViewFieldContent(defaultStringer),

		formTemplate:      "form_input_textarea",
		formValueStringer: stringerString,

		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{
		dbFieldDescription: "bigint(20)",
		getViewFieldContent: func(request *Request, field *Field, val any) *viewFieldContent {
			return &viewFieldContent{
				Name: humanizeNumber(val.(int64)) + ".",
			}
		},

		formValueStringer: stringerInt64,

		listCellDataSource: basicCellDataSource(numberStringer),

		formTemplate: "form_input_int",
	})

	app.addFieldType("cdnfile", &fieldType{
		dbFieldDescription: "varchar(255)",
		getViewFieldContent: func(request *Request, field *Field, val any) *viewFieldContent {
			return &viewFieldContent{
				CDNFileData: getCDNViewData(app, val.(string)),
			}
		},
		formTemplate:       "form_input_cdnfile",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",
		formValueStringer:    stringerString,
	})

	app.addFieldType("file", &fieldType{
		dbFieldDescription:  "varchar(255)",
		getViewFieldContent: imagePickerViewFieldContent,

		formTemplate:      "form_input_image",
		formDataSource:    imageFormDataSource(""),
		formValueStringer: stringerString,

		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		dbFieldDescription: "text",

		getViewFieldContent: imagePickerViewFieldContent,

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

		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			videoURL := filesCDN.GetVideoURL(value.(string))
			if videoURL == "" {
				return nil
			}
			return &viewFieldContent{
				VideoURL: videoURL,
			}
		},

		formTemplate:      "form_input",
		formValueStringer: stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),

		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		dbFieldDescription: "text",

		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			content := template.HTML(markdown.New(markdown.Breaks(true), markdown.HTML(true), markdown.Tables(true)).RenderToString([]byte(value.(string))))
			if content == "" {
				return nil
			}

			return &viewFieldContent{
				ContentHTML: template.HTML(content),
			}
		},
		formTemplate:      "form_input_markdown",
		formValueStringer: stringerString,

		listCellDataSource: markdownListDataSource,
	})
	app.addFieldType("place", &fieldType{
		dbFieldDescription: "varchar(255)",

		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			placeData := value.(string)
			if placeData == "" {
				return nil
			}
			return &viewFieldContent{
				PlaceData: placeData,
			}
		},

		formValueStringer: stringerString,

		listCellDataSource: basicCellDataSource(defaultStringer),

		formTemplate: "form_input_place",
	})

	app.addFieldType("relation", &fieldType{
		dbFieldDescription: "bigint(20)",
		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			valInt := value.(int64)
			previews := field.relationPreview(request, fmt.Sprintf("%d", valInt))
			if len(previews) == 0 {
				return nil
			}
			return &viewFieldContent{
				Previews: previews,
			}
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
		getViewFieldContent: func(request *Request, field *Field, value any) *viewFieldContent {
			previews := field.relationPreview(request, value.(string))
			if len(previews) == 0 {
				return nil
			}
			return &viewFieldContent{
				Previews: previews,
			}
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
		dbFieldDescription:  "date",
		getViewFieldContent: stringerToViewFieldContent(dateStringer),

		formTemplate:      "form_input_date",
		formValueStringer: stringerDate,
		naturalCellWidth:  130,

		listCellDataSource: basicCellDataSource(dateStringer),
	})

	app.addFieldType("time", &fieldType{
		dbFieldDescription:  "datetime",
		getViewFieldContent: stringerToViewFieldContent(timeStringer),

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
