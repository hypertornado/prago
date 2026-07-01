package prago

import (
	"fmt"
	"time"
)

// FieldType defines type of field
type fieldType struct {
	id string

	getViewFieldContent func(request *Request, val any) *viewFieldContent

	viewTemplate   string
	viewDataSource func(*Request, *Field, any) any

	dbFieldDescription string

	allowedValues []string

	formHideLabel   bool
	formTemplate    string
	formDataSource  func(*Field, UserData, string) any
	ft_formStringer func(any) string

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

	app.addFieldType("text", &fieldType{
		viewTemplate: "view_textarea",

		formTemplate: "form_input_textarea",

		listCellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{})
	app.addFieldType("date", &fieldType{
		naturalCellWidth: 130,
	})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:       "view_cdn_file",
		viewDataSource:     cdnViewDataSource,
		formTemplate:       "form_input_cdnfile",
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate: "filter_layout_text",
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:       "view_image",
		viewDataSource:     fileViewDataSource,
		formTemplate:       "form_input_image",
		formDataSource:     imageFormDataSource(""),
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
		listCellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,

		naturalCellWidth: 60,
	})
	app.addFieldType("video", &fieldType{
		viewTemplate:     "view_video",
		viewDataSource:   videoViewDataSource,
		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:       "view_markdown",
		viewDataSource:     markdownViewDataSource,
		formTemplate:       "form_input_markdown",
		listCellDataSource: markdownListDataSource,
	})
	app.addFieldType("place", &fieldType{
		viewTemplate: "view_place",
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
	})

	app.addFieldType("timestamp", &fieldType{
		formTemplate: "form_input_timestamp",
		ft_formStringer: func(i any) string {
			tm := i.(time.Time)
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02 15:04")
		},
		naturalCellWidth: 130,
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
