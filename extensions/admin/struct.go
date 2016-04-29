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
		"prago-admin-visible",
		"prago-admin-editable",
		"prago-preview",
		"prago-unique",
	} {
		ret.tags[v] = field.Tag.Get(v)
	}

	if ret.tags["prago-unique"] == "true" {
		ret.unique = true
	}

	return ret
}

func (cache *AdminStructCache) GetFormItemsDefault(ar *AdminResource, item interface{}, lang string) (*Form, error) {
	form := NewForm()

	form.Method = "POST"
	itemVal := reflect.ValueOf(item).Elem()

	for i, field := range cache.fieldArrays {

		visible := true
		var ifaceVal interface{}

		item := &FormItem{
			Name:        field.name,
			NameHuman:   field.name,
			SubTemplate: "admin_item_input",
		}

		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

		switch field.typ.Kind() {
		case reflect.Struct:
			if field.typ == reflect.TypeOf(time.Now()) {
				tm := ifaceVal.(time.Time)
				if field.tags["prago-admin-type"] == "timestamp" {
					item.SubTemplate = "admin_item_timestamp"
					item.Value = tm.Format("2006-01-02 15:04")
				} else {
					item.SubTemplate = "admin_item_date"
					item.Value = tm.Format("2006-01-02")
				}
				if item.Name == "CreatedAt" || item.Name == "UpdatedAt" {
					item.Readonly = true
				}
			} else {
				visible = false
			}
		case reflect.Bool:
			item.SubTemplate = "admin_item_checkbox"
			if ifaceVal.(bool) {
				item.Value = "on"
			}
			item.HiddenName = true
		case reflect.String:
			item.Value = ifaceVal.(string)
			switch field.tags["prago-admin-type"] {
			case "text":
				item.SubTemplate = "admin_item_textarea"
			case "image":
				item.SubTemplate = "admin_item_image"
			}
		case reflect.Int64:
			item.Value = fmt.Sprintf("%d", ifaceVal.(int64))
			if item.Name == "ID" {
				visible = false
			}
		default:
			visible = false
			panic("Wrong type" + field.typ.Kind().String())
		}

		item.NameHuman = field.humanName(lang)

		visibleTag := field.tags["prago-admin-visible"]
		if visibleTag == "true" {
			visible = true
		}
		if visibleTag == "false" {
			visible = false
		}

		editableTag := field.tags["prago-admin-editable"]
		if editableTag == "true" {
			item.Readonly = false
		}
		if editableTag == "false" {
			item.Readonly = true
		}

		if visible {
			form.AddItem(item)
		}
	}

	form.AddSubmit("_submit", "Submit")
	return form, nil
}
