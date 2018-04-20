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
}

func newField(f reflect.StructField, order int) *field {
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

	return ret
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
