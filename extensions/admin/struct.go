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

type StructCache struct {
	typ         reflect.Type
	fieldArrays []*StructField
	fieldMap    map[string]*StructField
}

func NewStructCache(item interface{}) (ret *StructCache, err error) {
	typ := reflect.TypeOf(item)
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("item is not a structure")
	}

	ret = &StructCache{
		typ:      typ,
		fieldMap: make(map[string]*StructField),
	}

	for i := 0; i < typ.NumField(); i++ {
		field := newStructField(typ.Field(i), i)
		ret.fieldArrays = append(ret.fieldArrays, field)
		ret.fieldMap[field.Name] = field
	}

	return
}

type StructField struct {
	Name          string
	LowercaseName string
	Typ           reflect.Type
	Tags          map[string]string
	Order         int
	Unique        bool
	Scanner       sql.Scanner
}

func (a *StructField) fieldDescriptionMysql() string {
	var fieldDescription string
	switch a.Typ.Kind() {
	case reflect.Struct:
		dateType := reflect.TypeOf(time.Now())
		if a.Typ == dateType {
			fieldDescription = "datetime"
		}
	case reflect.Bool:
		fieldDescription = "bool"
	case reflect.Int64:
		fieldDescription = "bigint(20)"
	case reflect.String:
		if a.Tags["prago-type"] == "text" {
			fieldDescription = "text"
		} else {
			fieldDescription = "varchar(255)"
		}
	}

	additional := ""
	if a.LowercaseName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if a.Unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", a.LowercaseName, fieldDescription, additional)
}

func (a *StructField) humanName(lang string) (ret string) {
	description := a.Tags["prago-description"]
	if len(description) > 0 {
		return description
	} else {
		translatedName := messages.Messages.GetNullable(lang, a.Name)
		if translatedName == nil {
			return a.Name
		} else {
			return *translatedName
		}
	}
}

func newStructField(field reflect.StructField, order int) *StructField {
	ret := &StructField{
		Name:          field.Name,
		LowercaseName: utils.PrettyUrl(field.Name),
		Typ:           field.Type,
		Tags:          make(map[string]string),
		Order:         order,
	}

	for _, v := range []string{
		"prago-type",
		"prago-description",
		"prago-visible",
		"prago-editable",
		"prago-preview",
		"prago-unique",
	} {
		ret.Tags[v] = field.Tag.Get(v)
	}

	if ret.Tags["prago-unique"] == "true" {
		ret.Unique = true
	}

	return ret
}

type StructFieldVisibler func(field *StructField) bool

func DefaultVisibler(field *StructField) bool {
	visible := true
	//if field.name
	return visible
}

func (cache *StructCache) GetFormItemsDefault(inValues interface{}, lang string) (*Form, error) {
	form := NewForm()

	form.Method = "POST"
	itemVal := reflect.ValueOf(inValues).Elem()

	for i, field := range cache.fieldArrays {

		visible := true
		var ifaceVal interface{}

		item := &FormItem{
			Name:        field.Name,
			NameHuman:   field.Name,
			SubTemplate: "admin_item_input",
		}

		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

		switch field.Typ.Kind() {
		case reflect.Struct:
			if field.Typ == reflect.TypeOf(time.Now()) {
				tm := ifaceVal.(time.Time)
				if field.Tags["prago-type"] == "timestamp" {
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
			switch field.Tags["prago-type"] {
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
			panic("Wrong type" + field.Typ.Kind().String())
		}

		item.NameHuman = field.humanName(lang)

		visibleTag := field.Tags["prago-visible"]
		if visibleTag == "true" {
			visible = true
		}
		if visibleTag == "false" {
			visible = false
		}

		editableTag := field.Tags["prago-editable"]
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

	return form, nil
}
