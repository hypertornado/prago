package administration

import (
	"reflect"
	"time"
)

type FieldType struct {
	ViewTemplate       string
	FormSubTemplate    string
	DBFieldDescription string
	ValuesSource       *func() interface{}
}

var defaultFieldType = FieldType{
	ViewTemplate:       "admin_item_view_text",
	FormSubTemplate:    "admin_item_input",
	DBFieldDescription: "varchar(255)",
	ValuesSource:       nil,
}

var stringFieldType = FieldType{
	DBFieldDescription: "varchar(255)",
}

var boolFieldType = FieldType{
	FormSubTemplate:    "admin_item_checkbox",
	DBFieldDescription: "bool NOT NULL",
}

var intFieldType = FieldType{
	DBFieldDescription: "bigint(20)",
}

var floatFieldType = FieldType{
	DBFieldDescription: "double",
}

var dateFieldType = FieldType{
	FormSubTemplate:    "admin_item_date",
	DBFieldDescription: "datetime",
}

func defaultFieldTypes() map[string]FieldType {
	ret := map[string]FieldType{}
	return ret
}

func getDefaultField(in reflect.Type) FieldType {
	switch in.Kind() {
	case reflect.String:
		return stringFieldType
	case reflect.Bool:
		return boolFieldType
	case reflect.Int, reflect.Int64:
		return intFieldType
	case reflect.Float64:
		return floatFieldType
	case reflect.Struct:
		switch in {
		case reflect.TypeOf(time.Now()):
			return dateFieldType
		}
	}

	return FieldType{}
}
