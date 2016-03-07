package extensions

import (
	"database/sql"
	"errors"
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
		field := newAdminStructField(typ.Field(i), i)
		ret.fieldArrays = append(ret.fieldArrays, field)
		ret.fieldMap[field.name] = field
	}

	return
}

type adminStructField struct {
	name             string
	lowercaseName    string
	typ              reflect.Type
	tags             map[string]string
	order            int
	mysqlDescription string
	scanner          sql.Scanner
}

func newAdminStructField(field reflect.StructField, order int) *adminStructField {
	ret := &adminStructField{
		name:          field.Name,
		lowercaseName: utils.PrettyUrl(field.Name),
		typ:           field.Type,
		tags:          make(map[string]string),
		order:         order,
	}

	for _, v := range []string{"prago-admin-type", "prago-admin-description", "prago-admin-access"} {
		ret.tags[v] = field.Tag.Get(v)
	}

	ret.mysqlDescription = ret.getMysqlDescription()

	return ret
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

func (cache *AdminStructCache) GetFormItemsDefault(ar *AdminResource, item interface{}) ([]AdminFormItem, error) {
	itemVal := reflect.ValueOf(item).Elem()
	items := []AdminFormItem{}

	for i, field := range cache.fieldArrays {
		structItem := AdminFormItem{
			Name:      field.name,
			NameHuman: field.name,
			Template:  "admin_item_input",
		}

		reflect.ValueOf(&structItem.Value).Elem().Set(
			itemVal.Field(i),
		)

		switch field.typ.Kind() {
		case reflect.Struct:
			if field.typ == reflect.TypeOf(time.Now()) {
				structItem.Template = "admin_item_date"
				var tm time.Time
				reflect.ValueOf(&tm).Elem().Set(reflect.ValueOf(structItem.Value))
				newVal := reflect.New(reflect.TypeOf("")).Elem()
				newVal.SetString(tm.Format("2006-01-02"))
				reflect.ValueOf(&structItem.Value).Elem().Set(newVal)
			}
		case reflect.Bool:
			structItem.Template = "admin_item_checkbox"
		case reflect.String:
			switch field.tags["prago-admin-type"] {
			case "text":
				structItem.Template = "admin_item_textarea"
			case "image":
				structItem.Template = "admin_item_image"
			}
		}

		description := field.tags["prago-admin-description"]
		if len(description) > 0 {
			structItem.NameHuman = description
		}

		accessTag := field.tags["prago-admin-access"]
		if accessTag == "-" || structItem.Name == "CreatedAt" || structItem.Name == "UpdatedAt" {
			structItem.Template = "admin_item_readonly"
		}

		if structItem.Name != "ID" {
			items = append(items, structItem)
		}
	}
	return items, nil
}
