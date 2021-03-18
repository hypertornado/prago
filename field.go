package prago

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type field struct {
	Name       string
	ColumnName string
	HumanName  func(string) string
	Typ        reflect.Type
	Tags       map[string]string
	fieldOrder int
	Unique     bool
	Scanner    sql.Scanner
	CanOrder   bool

	resource  *Resource
	fieldType FieldType
}

//GetRelatedResourceName gets related resource name
func (field field) GetRelatedResourceName() string {
	relatedTag := field.Tags["prago-relation"]
	if relatedTag != "" {
		return strings.ToLower(relatedTag)
	}
	return field.ColumnName
}

func (resource *Resource) newField(f reflect.StructField, order int, fieldTypes map[string]FieldType) *field {
	ret := &field{
		Name:       f.Name,
		ColumnName: columnName(f.Name),
		HumanName:  Unlocalized(f.Name),
		Typ:        f.Type,
		Tags:       make(map[string]string),
		fieldOrder: order,
		CanOrder:   true,

		resource: resource,
	}

	for _, v := range []string{
		"prago-edit",
		"prago-view",

		"prago-type",
		"prago-description",
		"prago-visible",
		"prago-editable",
		"prago-preview",
		"prago-unique",
		"prago-order",
		"prago-order-desc",
		"prago-relation",
	} {
		ret.Tags[v] = f.Tag.Get(v)
	}

	if ret.Tags["prago-unique"] == "true" {
		ret.Unique = true
	}

	description := ret.Tags["prago-description"]
	if len(description) > 0 {
		ret.HumanName = Unlocalized(description)
	} else {
		messages.GetNameFunction(ret.Name)
		nameFunction := messages.GetNameFunction(ret.Name)
		if nameFunction != nil {
			ret.HumanName = nameFunction
		}
	}

	ret.initFieldType(fieldTypes)

	return ret
}

/*
if ret.Tags["prago-type"] == "relation" {
			relationID := ret.Name
			ret.Tags["prago-relation"] != "" {
				relationID = ret.Tags["prago-relation"]
			}
			referenceResource
		}

*/

func getDefaultStringer(t reflect.Type) func(interface{}) string {
	if reflect.TypeOf(time.Now()) == t {
		return func(i interface{}) string {
			tm := i.(time.Time)
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02")
		}
	}

	switch t.Kind() {
	case reflect.String:
		return func(i interface{}) string {
			return i.(string)
		}
	case reflect.Int64:
		return func(i interface{}) string {
			return fmt.Sprintf("%d", i.(int64))
		}
	case reflect.Float64:
		return func(i interface{}) string {
			return fmt.Sprintf("%f", i.(float64))
		}
	case reflect.Bool:
		return func(i interface{}) string {
			if i.(bool) {
				return "on"
			}
			return ""
		}
	}
	panic("unknown stringer for " + t.String())
}

func getDefaultFormTemplate(t reflect.Type) string {
	if t == reflect.TypeOf(time.Now()) {
		return "admin_item_date"
	}

	switch t.Kind() {
	case reflect.String:
		return "admin_item_input"
	case reflect.Bool:
		return "admin_item_checkbox"
	case reflect.Int64:
		return "admin_item_input_int"
	case reflect.Float64:
		return "admin_item_input_float"
	}
	panic("unknown default form for " + t.String())
}

func (field *field) initFieldType(fieldTypes map[string]FieldType) {
	fieldTypeName := field.Tags["prago-type"]

	ret, found := fieldTypes[fieldTypeName]
	if !found && fieldTypeName != "" {
		panic(fmt.Sprintf("Field type '%s' not found", fieldTypeName))
	}

	if ret.ViewTemplate == "" {
		ret.ViewTemplate = getDefaultViewTemplate(field.Typ)
	}
	if ret.ViewDataSource == nil {
		ret.ViewDataSource = getDefaultViewDataSource(field)
	}

	if ret.FormTemplate == "" {
		ret.FormTemplate = getDefaultFormTemplate(field.Typ)
	}

	if ret.FormStringer == nil {
		ret.FormStringer = getDefaultStringer(field.Typ)
	}

	if ret.FormTemplate == "admin_item_checkbox" {
		ret.FormHideLabel = true
	}

	if ret.ListCellDataSource == nil {
		ret.ListCellDataSource = ret.ViewDataSource
	}
	if ret.ListCellTemplate == "" {
		ret.ListCellTemplate = ret.ViewTemplate
	}

	field.fieldType = ret
}

func (field field) fieldDescriptionMysql(fieldTypes map[string]FieldType) string {
	var fieldDescription string

	t, found := fieldTypes[field.Tags["prago-type"]]
	if found && t.DBFieldDescription != "" {
		fieldDescription = t.DBFieldDescription
	} else {
		switch field.Typ.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if field.Typ == dateType {
				if field.Tags["prago-type"] == "date" {
					fieldDescription = "date"
				} else {
					fieldDescription = "datetime"
				}
			}
		case reflect.Bool:
			fieldDescription = "bool NOT NULL"
		case reflect.Float64:
			fieldDescription = "double"
		case reflect.Int64:
			fieldDescription = "bigint(20)"
		case reflect.String:
			if field.Tags["prago-type"] == "text" || field.Tags["prago-type"] == "image" || field.Tags["prago-type"] == "markdown" {
				fieldDescription = "text"
			} else {
				fieldDescription = "varchar(255)"
			}
		default:
			panic("non supported type " + field.Typ.Kind().String())
		}
	}

	additional := ""
	if field.ColumnName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if field.Unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", field.ColumnName, fieldDescription, additional)
}

func (field field) shouldShow() (show bool) {
	if field.Name == "Name" {
		show = true
	}
	showTag := field.Tags["prago-preview"]
	if showTag == "true" {
		show = true
	}
	if showTag == "false" {
		show = false
	}
	return
}

func (field field) getRelatedResource(app App) *Resource {
	if field.Tags["prago-type"] != "relation" {
		return nil
	}
	var relationName string
	if field.Tags["prago-relation"] != "" {
		relationName = field.Tags["prago-relation"]
	} else {
		relationName = field.Name
	}
	return app.getResourceByName(relationName)
}
