package administration

import (
	"database/sql"
	"fmt"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
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

	fieldType FieldType

	/*FormHideLabel  bool
	FormTemplate   string
	FormDataSource *func(User) interface{}
	FormStringer   func(interface{}) string*/
}

func newField(f reflect.StructField, order int, fieldTypes map[string]FieldType) *field {
	ret := &field{
		Name:       f.Name,
		ColumnName: columnName(f.Name),
		HumanName:  Unlocalized(f.Name),
		Typ:        f.Type,
		Tags:       make(map[string]string),
		fieldOrder: order,
		CanOrder:   true,
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
		"prago-preview-type",
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
		messages.Messages.GetNameFunction(ret.Name)
		nameFunction := messages.Messages.GetNameFunction(ret.Name)
		if nameFunction != nil {
			ret.HumanName = nameFunction
		}
	}

	ret.initFieldType(fieldTypes)

	return ret
}

func getDefaultStringer(t reflect.Type) func(interface{}) string {
	if reflect.TypeOf(time.Now()) == t {
		return func(i interface{}) string {
			return i.(time.Time).Format("2006-01-02")
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
		return "admin_item_input"
	case reflect.Float64:
		return "admin_item_input"
	}
	panic("unknown default form for " + t.String())
}

func (f *field) initFieldType(fieldTypes map[string]FieldType) {
	ret := fieldTypes[f.Tags["prago-type"]]

	if ret.FormTemplate == "" {
		ret.FormTemplate = getDefaultFormTemplate(f.Typ)
	}

	if ret.FormStringer == nil {
		ret.FormStringer = getDefaultStringer(f.Typ)
	}

	if ret.FormTemplate == "admin_item_checkbox" {
		ret.FormHideLabel = true
	}

	f.fieldType = ret

	/*
		switch f.Typ.Kind() {
		case reflect.Struct:
			if f.Typ == reflect.TypeOf(time.Now()) {
				tm := ifaceVal.(time.Time)
				if f.Tags["prago-type"] == "timestamp" {
					item.SubTemplate = "admin_item_timestamp"
					item.Value = tm.Format("2006-01-02 15:04")
				} else {
					item.SubTemplate = "admin_item_date"
					item.Value = tm.Format("2006-01-02")
				}
			}
		case reflect.Bool:
			item.SubTemplate = "admin_item_checkbox"
			if ifaceVal.(bool) {
				item.Value = "on"
			}
			item.HiddenName = true
		case reflect.String:
			item.Value = ifaceVal.(string)
			switch f.Tags["prago-type"] {
			case "text":
				item.SubTemplate = "admin_item_textarea"
			case "markdown":
				item.Template = "admin_item_markdown"
			case "image":
				item.SubTemplate = "admin_item_image"
			case "place":
				item.SubTemplate = "admin_item_place"
			}
		case reflect.Int64:
			item.Value = fmt.Sprintf("%d", ifaceVal.(int64))
			switch f.Tags["prago-type"] {
			case "relation":
				item.SubTemplate = "admin_item_relation"
				if f.Tags["prago-relation"] != "" {
					item.Data = columnName(f.Tags["prago-relation"])
				} else {
					item.Data = columnName(item.Name)
				}
			}
		case reflect.Float64:
			item.Value = fmt.Sprintf("%f", ifaceVal.(float64))
		default:
			panic("Wrong type" + f.Typ.Kind().String())
		}
	*/
}

func (admin *Administration) addDefaultFieldTypes() {
	admin.AddFieldType("role", admin.createRoleFieldType())

	admin.AddFieldType("timestamp", FieldType{FormTemplate: "admin_item_textarea"})
	admin.AddFieldType("markdown", FieldType{FormTemplate: "admin_item_markdown"})
	admin.AddFieldType("image", FieldType{FormTemplate: "admin_item_image"})
	admin.AddFieldType("place", FieldType{FormTemplate: "admin_item_place"})

	admin.AddFieldType("relation", FieldType{
		FormTemplate: "admin_item_relation",
		FormDataSource: func(f field, u User) interface{} {
			if f.Tags["prago-relation"] != "" {
				return columnName(f.Tags["prago-relation"])
			} else {
				return columnName(f.Name)
			}
		},
	})

	admin.AddFieldType("timestamp", FieldType{
		FormTemplate: "admin_item_timestamp",
		FormStringer: func(i interface{}) string {
			return i.(time.Time).Format("2006-01-02 15:04")
		},
	})
}

func (sf field) fieldDescriptionMysql(fieldTypes map[string]FieldType) string {
	var fieldDescription string

	t, found := fieldTypes[sf.Tags["prago-type"]]
	if found && t.DBFieldDescription != "" {
		fieldDescription = t.DBFieldDescription
	} else {
		switch sf.Typ.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if sf.Typ == dateType {
				if sf.Tags["prago-type"] == "date" {
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
			if sf.Tags["prago-type"] == "text" || sf.Tags["prago-type"] == "image" || sf.Tags["prago-type"] == "markdown" {
				fieldDescription = "text"
			} else {
				fieldDescription = "varchar(255)"
			}
		default:
			panic("non supported type " + sf.Typ.Kind().String())
		}
	}

	additional := ""
	if sf.ColumnName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if sf.Unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", sf.ColumnName, fieldDescription, additional)
}

func (sf field) shouldShow() (show bool) {
	if sf.Name == "Name" {
		show = true
	}
	showTag := sf.Tags["prago-preview"]
	if showTag == "true" {
		show = true
	}
	if showTag == "false" {
		show = false
	}
	return
}
