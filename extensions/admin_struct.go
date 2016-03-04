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
		tags:          make(map[string]string),
	}

	for _, v := range []string{"prago-admin-type"} {
		ret.tags[v] = field.Tag.Get(v)
	}

	ret.mysqlDescription = ret.getMysqlDescription()

	return ret
}

type adminStructField struct {
	name             string
	lowercaseName    string
	typ              reflect.Type
	tags             map[string]string
	mysqlDescription string
}

func (f *adminStructField) getMysqlDescription() string {
	switch f.typ.Kind() {
	case reflect.Struct:
		dateType := reflect.TypeOf(time.Now())
		if f.typ == dateType {
			return "datetime"
		}
	case reflect.Bool:
		return "bool"
	case reflect.Int64:
		return "bigint(20)"
	case reflect.String:
		if f.tags["prago-admin-type"] == "text" {
			return "text"
		} else {
			return "varchar(255)"
		}
	}
	return ""
}

func (s *AdminStructCache) getStructDescription() (columns []*mysqlColumn, err error) {
	for _, field := range s.fieldArrays {
		if len(field.mysqlDescription) > 0 {
			columns = append(columns, &mysqlColumn{
				Field: field.lowercaseName,
				Type:  field.mysqlDescription,
			})
		}
	}
	return
}

func (s *AdminStructCache) getStructScanners(value reflect.Value) (names []string, scanners []interface{}, err error) {

	for i := 0; i < value.Type().NumField(); i++ {
		use := true
		field := value.Type().Field(i)
		name := utils.PrettyUrl(field.Name)

		switch field.Type.Kind() {
		case reflect.Int64:
		case reflect.Bool:
		case reflect.String:
		case reflect.Struct:
			if field.Type != reflect.TypeOf(time.Now()) {
				use = false
			}
		default:
			use = false
		}
		if use {
			names = append(names, name)
			scanners = append(scanners, &scanner{value.Field(i)})
		}
	}
	return
}
