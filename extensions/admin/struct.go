package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago/extensions/admin/messages"
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
	name          string
	lowercaseName string
	typ           reflect.Type
	tags          map[string]string
	order         int
	unique        bool
	scanner       sql.Scanner
}

func (a *adminStructField) fieldDescriptionMysql() string {
	var fieldDescription string
	switch a.typ.Kind() {
	case reflect.Struct:
		dateType := reflect.TypeOf(time.Now())
		if a.typ == dateType {
			fieldDescription = "datetime"
		}
	case reflect.Bool:
		fieldDescription = "bool"
	case reflect.Int64:
		fieldDescription = "bigint(20)"
	case reflect.String:
		if a.tags["prago-admin-type"] == "text" {
			fieldDescription = "text"
		} else {
			fieldDescription = "varchar(255)"
		}
	}

	additional := ""
	if a.lowercaseName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if a.unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", a.lowercaseName, fieldDescription, additional)
}

func (a *adminStructField) humanName(lang string) (ret string) {
	description := a.tags["prago-admin-description"]
	if len(description) > 0 {
		return description
	} else {
		translatedName := messages.Messages.GetNullable(lang, a.name)
		if translatedName == nil {
			return a.name
		} else {
			return *translatedName
		}
	}
}

func newAdminStructField(field reflect.StructField, order int) *adminStructField {
	ret := &adminStructField{
		name:          field.Name,
		lowercaseName: utils.PrettyUrl(field.Name),
		typ:           field.Type,
		tags:          make(map[string]string),
		order:         order,
	}

	for _, v := range []string{
		"prago-admin-type",
		"prago-admin-description",
		"prago-admin-access",
		"prago-admin-show",
		"prago-admin-unique",
	} {
		ret.tags[v] = field.Tag.Get(v)
	}

	if ret.tags["prago-admin-unique"] == "true" {
		ret.unique = true
	}

	return ret
}

func (cache *AdminStructCache) GetFormItemsDefault(ar *AdminResource, item interface{}, lang string) ([]AdminFormItem, error) {
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
				var tm time.Time
				reflect.ValueOf(&tm).Elem().Set(reflect.ValueOf(structItem.Value))
				newVal := reflect.New(reflect.TypeOf("")).Elem()

				if field.tags["prago-admin-type"] == "timestamp" {
					structItem.Template = "admin_item_timestamp"
					newVal.SetString(tm.Format("2006-01-02 15:04"))
				} else {
					structItem.Template = "admin_item_date"
					newVal.SetString(tm.Format("2006-01-02"))
				}
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

		structItem.NameHuman = field.humanName(lang)

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
