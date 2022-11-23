package prago

import (
	"fmt"
	"strings"
	"time"
)

// FieldType defines type of field
type fieldType struct {
	viewTemplate   string
	viewDataSource func(*user, *Field, interface{}) interface{}

	dbFieldDescription string

	allowedValues []string

	formHideLabel  bool
	formTemplate   string
	formDataSource func(*Field, *user) interface{}
	formStringer   func(interface{}) string

	cellDataSource func(*user, *Field, interface{}) cellViewData

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
		viewTemplate:   "admin_item_view_textarea",
		formTemplate:   "admin_item_textarea",
		cellDataSource: textListDataSource,
	})
	app.addFieldType("order", &fieldType{})
	app.addFieldType("date", &fieldType{
		naturalCellWidth: 130,
	})

	app.addFieldType("cdnfile", &fieldType{
		viewTemplate:   "admin_item_view_file",
		viewDataSource: filesViewDataSource,
		formTemplate:   "admin_file",
		cellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
	})

	app.addFieldType("file", &fieldType{
		viewTemplate:   "admin_item_view_image",
		formTemplate:   "admin_item_image",
		formDataSource: createFilesEditDataSource(""),
		cellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,
		naturalCellWidth:       60,
	})

	app.addFieldType("image", &fieldType{
		viewTemplate:   "admin_item_view_image",
		formTemplate:   "admin_item_image",
		formDataSource: createFilesEditDataSource(".jpg,.jpeg,.png"),
		cellDataSource: imageCellViewData,

		filterLayoutTemplate:   "filter_layout_select",
		filterLayoutDataSource: boolFilterLayoutDataSource,

		fieldTypeIcon:    "glyphicons-basic-38-picture.svg",
		naturalCellWidth: 60,
	})

	app.addFieldType("markdown", &fieldType{
		viewTemplate:   "admin_item_view_markdown",
		viewDataSource: markdownViewDataSource,
		formTemplate:   "admin_item_markdown",
		cellDataSource: markdownListDataSource,
		//listCellTemplate:   "admin_item_view_text",
	})
	app.addFieldType("place", &fieldType{
		viewTemplate: "admin_item_view_place",
		formTemplate: "admin_item_place",
		//listCellTemplate: "admin_item_view_text",
	})

	app.addFieldType("relation", &fieldType{
		viewTemplate: "admin_item_view_relation",
		//listCellTemplate: "admin_item_view_relation_cell",
		viewDataSource: getRelationViewData,
		formTemplate:   "admin_item_relation",
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

func textListDataSource(user *user, f *Field, value interface{}) cellViewData {
	return cellViewData{Name: crop(value.(string), 100)}
}

func createFilesEditDataSource(mimeTypes string) func(f *Field, u *user) interface{} {
	return func(f *Field, u *user) interface{} {
		return mimeTypes
	}
}

func markdownViewDataSource(user *user, f *Field, value interface{}) interface{} {
	return cropMarkdown(value.(string), 100)
}

func markdownListDataSource(user *user, f *Field, value interface{}) cellViewData {
	return cellViewData{Name: cropMarkdown(value.(string), 100)}
}

func relationCellViewData(user *user, f *Field, value interface{}) cellViewData {
	previewData, _ := getPreviewData(user, f, value.(int64))
	if previewData == nil {
		return cellViewData{}
	}

	ret := cellViewData{
		Name: previewData.Name,
	}
	if previewData.Image != "" {
		ret.Images = []string{previewData.Image}
	}
	return ret
}

func imageCellViewData(user *user, f *Field, value interface{}) cellViewData {

	data := value.(string)
	images := strings.Split(data, ",")
	ret := cellViewData{}
	if len(images) > 0 {
		ret.Images = images
	}
	return ret

	/*previewData, _ := getPreviewData(user, f, value.(int64))
	ret := cellViewData{
		Name: previewData.Name,
	}
	if previewData.Image != "" {
		ret.Images = []string{previewData.Image}
	}
	return ret*/
}

type cellViewData struct {
	Images []string
	Name   string
}

func getCellViewData(user *user, f *Field, value interface{}) cellViewData {
	if f.fieldType.cellDataSource != nil {
		return f.fieldType.cellDataSource(user, f, value)
	}

	if f.fieldType.IsRelation() {
		return relationCellViewData(user, f, value)
	}

	ret := cellViewData{
		Name: getDefaultFieldStringer(f)(user, f, value),
	}
	return ret

}
