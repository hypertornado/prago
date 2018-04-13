package administration

import (
	"fmt"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
	"time"
)

type view struct {
	Items []viewData
}

type viewData struct {
	Name     string
	Template string
	Value    interface{}
}

type viewRelationData struct {
	Typ string
	ID  int64
}

func (cache *structCache) getView(inValues interface{}, lang string, visible structFieldFilter, editable structFieldFilter) (*view, error) {
	ret := view{}

	itemVal := reflect.ValueOf(inValues).Elem()

	for i, field := range cache.fieldArrays {
		if !visible(field) {
			continue
		}

		var ifaceVal interface{}

		reflect.ValueOf(&ifaceVal).Elem().Set(
			itemVal.Field(i),
		)

		item := viewData{
			Name:     field.Name,
			Template: "admin_item_view_text",
			Value:    ifaceVal,
		}

		t, found := cache.fieldTypes[field.Tags["prago-type"]]
		if found && t.ViewTemplate != "" {
			item.Template = t.ViewTemplate
		} else {
			switch field.Typ.Kind() {
			case reflect.Struct:
				if field.Typ == reflect.TypeOf(time.Now()) {
					tm := ifaceVal.(time.Time)
					if field.Tags["prago-type"] == "timestamp" {
						item.Value = tm.Format("2006-01-02 15:04")
					} else {
						item.Value = tm.Format("2006-01-02")
					}
				}
			case reflect.Bool:
				if ifaceVal.(bool) {
					item.Value = messages.Messages.Get(lang, "yes")
				} else {
					item.Value = messages.Messages.Get(lang, "no")
				}
			case reflect.String:
				switch field.Tags["prago-type"] {
				case "markdown":
					item.Template = "admin_item_view_markdown"
				case "image":
					item.Template = "admin_item_view_image"
				case "place":
					item.Template = "admin_item_view_place"
				}
			case reflect.Int64:
				switch field.Tags["prago-type"] {
				case "relation":
					item.Template = "admin_item_view_relation"
					var val = viewRelationData{}
					if field.Tags["prago-relation"] != "" {
						val.Typ = columnName(field.Tags["prago-relation"])
					} else {
						val.Typ = columnName(item.Name)
					}
					val.ID = ifaceVal.(int64)
					item.Value = val
				}
			case reflect.Float64:
				item.Value = fmt.Sprintf("%f", ifaceVal.(float64))
			default:
				panic("Wrong type" + field.Typ.Kind().String())
			}
		}

		item.Name = field.humanName(lang)

		ret.Items = append(ret.Items, item)
	}

	return &ret, nil
}
