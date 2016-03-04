package extensions

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"reflect"
	"time"
)

type AdminStructCache struct {
	typ         reflect.Type
	fieldArrays []*adminStructField
	fieldMap    map[string]*adminStructField
}

func NewAdminStructCache(item interface{}) (ret *AdminStructCache, err error) {
	typ := reflect.TypeOf(item)
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("item is not a structure")
	}

	ret = &AdminStructCache{
		typ:      typ,
		fieldMap: make(map[string]*adminStructField),
	}

	for i := 0; i < typ.NumField(); i++ {
		field := newAdminStructField(typ.Field(i))
		ret.fieldArrays = append(ret.fieldArrays, field)
		ret.fieldMap[field.name] = field
	}

	fmt.Println(ret)

	return
}

func newAdminStructField(field reflect.StructField) *adminStructField {
	ret := &adminStructField{
		name:          field.Name,
		lowercaseName: utils.PrettyUrl(field.Name),
		typ:           field.Type,
	}
	return ret
}

type adminStructField struct {
	name          string
	lowercaseName string
	typ           reflect.Type
}

func (s *AdminStructCache) getStructDescription() (columns []*mysqlColumn, err error) {
	typ := s.typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		use := true
		field := typ.Field(i)
		column := &mysqlColumn{
			Field: utils.PrettyUrl(field.Name),
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if field.Type == dateType {
				column.Type = "datetime"
			} else {
				use = false
			}
		case reflect.Bool:
			column.Type = "bool"
		case reflect.Int64:
			column.Type = "bigint(20)"
		case reflect.String:
			if field.Tag.Get("prago-admin-type") == "text" {
				column.Type = "text"
			} else {
				column.Type = "varchar(255)"
			}
		default:
			use = false
		}
		if use {
			columns = append(columns, column)
		}
	}
	return
}
