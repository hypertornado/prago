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
	typ             reflect.Type
	fieldArrays     []*StructField
	fieldMap        map[string]*StructField
	OrderFieldName  string
	OrderColumnName string
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
		if field.Tags["prago-type"] == "order" {
			ret.OrderFieldName = field.Name
			ret.OrderColumnName = field.ColumnName
		}
		ret.fieldArrays = append(ret.fieldArrays, field)
		ret.fieldMap[field.Name] = field
	}
	return
}

func (cs *StructCache) GetDefaultOrder() (column string, desc bool) {
	column = ""

	for _, v := range cs.fieldArrays {
		add := false
		if v.ColumnName == "id" {
			add = true
		}
		if v.Tags["prago-type"] == "order" {
			add = true
		}
		if v.Tags["prago-order"] == "true" {
			add = true
		}
		if v.Tags["prago-order-desc"] == "true" {
			add = true
		}
		if v.Tags["prago-order"] == "false" {
			add = false
		}

		if add == true {
			column = v.ColumnName
			desc = false
			if v.Tags["prago-order-desc"] == "true" {
				desc = true
			}
		}
	}
	return
}

type StructField struct {
	Name       string
	ColumnName string
	Typ        reflect.Type
	Tags       map[string]string
	Order      int
	Unique     bool
	Scanner    sql.Scanner
}

func (a *StructField) fieldDescriptionMysql() string {
	var fieldDescription string
	switch a.Typ.Kind() {
	case reflect.Struct:
		dateType := reflect.TypeOf(time.Now())
		if a.Typ == dateType {
			if a.Tags["prago-type"] == "date" {
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
		if a.Tags["prago-type"] == "text" || a.Tags["prago-type"] == "image" || a.Tags["prago-type"] == "markdown" {
			fieldDescription = "text"
		} else {
			fieldDescription = "varchar(255)"
		}
	default:
		panic("non supported type " + a.Typ.Kind().String())
	}

	additional := ""
	if a.ColumnName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if a.Unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", a.ColumnName, fieldDescription, additional)
}

func (a *StructField) humanName(lang string) (ret string) {
	description := a.Tags["prago-description"]
	if len(description) > 0 {
		return description
	}
	translatedName := messages.Messages.GetNullable(lang, a.Name)
	if translatedName == nil {
		return a.Name
	}
	return *translatedName
}

func newStructField(field reflect.StructField, order int) *StructField {
	ret := &StructField{
		Name:       field.Name,
		ColumnName: utils.ColumnName(field.Name),
		Typ:        field.Type,
		Tags:       make(map[string]string),
		Order:      order,
	}

	for _, v := range []string{
		"prago-type",
		"prago-description",
		"prago-visible",
		"prago-editable",
		"prago-preview",
		"prago-unique",
		"prago-order",
		"prago-order-desc",
		"prago-preview-type",
	} {
		ret.Tags[v] = field.Tag.Get(v)
	}

	if ret.Tags["prago-unique"] == "true" {
		ret.Unique = true
	}

	return ret
}

type StructFieldFilter func(field *StructField) bool

func DefaultVisibilityFilter(field *StructField) bool {
	visible := true
	if field.Name == "ID" {
		visible = false
	}

	if field.Tags["prago-type"] == "order" {
		visible = false
	}

	visibleTag := field.Tags["prago-visible"]
	if visibleTag == "true" {
		visible = true
	}
	if visibleTag == "false" {
		visible = false
	}
	return visible
}

func DefaultEditabilityFilter(field *StructField) bool {
	editable := true
	if field.Name == "CreatedAt" || field.Name == "UpdatedAt" {
		editable = false
	}

	editableTag := field.Tags["prago-editable"]
	if editableTag == "true" {
		editable = true
	}
	if editableTag == "false" {
		editable = false
	}
	return editable
}

func WhiteListFilter(in ...string) StructFieldFilter {
	m := make(map[string]bool)
	for _, v := range in {
		m[v] = true
	}
	return func(field *StructField) bool {
		return m[field.Name]
	}
}

func (cache *StructCache) GetForm(inValues interface{}, lang string, visible StructFieldFilter, editable StructFieldFilter) (*Form, error) {
	form := NewForm()

	form.Method = "POST"
	itemVal := reflect.ValueOf(inValues).Elem()

	for i, field := range cache.fieldArrays {
		if !visible(field) {
			continue
		}

		var ifaceVal interface{}

		item := &FormItem{
			Name:        field.Name,
			NameHuman:   field.Name,
			SubTemplate: "admin_item_input",
		}

		if !editable(field) {
			item.Readonly = true
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
			case "markdown":
				item.SubTemplate = "admin_item_markdown"
			case "image":
				item.SubTemplate = "admin_item_image"
			case "place":
				item.SubTemplate = "admin_item_place"
			}
		case reflect.Int64:
			item.Value = fmt.Sprintf("%d", ifaceVal.(int64))
			switch field.Tags["prago-type"] {
			case "relation":
				item.SubTemplate = "admin_item_relation"
				item.Values = utils.ColumnName(item.Name)
			}
		case reflect.Float64:
			item.Value = fmt.Sprintf("%f", ifaceVal.(float64))
		default:
			panic("Wrong type" + field.Typ.Kind().String())
		}

		item.NameHuman = field.humanName(lang)

		form.AddItem(item)
	}

	return form, nil
}
