package administration

import (
	"fmt"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
	"time"
)

type view struct {
	Items []viewField
}

type viewField struct {
	Name     string
	Template string
	Value    interface{}
}

type viewRelationData struct {
	Typ string
	ID  int64
}

func (resource Resource) getView(inValues interface{}, user User, visible structFieldFilter) view {
	ret := view{}
	for i, field := range resource.StructCache.fieldArrays {
		if !visible(resource, user, *field) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(inValues).Elem().Field(i),
		)

		ret.Items = append(ret.Items,
			getViewField(resource.StructCache, user, *field, ifaceVal),
		)
	}
	return ret
}

func getViewField(cache *structCache, user User, field structField, ifaceVal interface{}) viewField {
	item := viewField{
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
				item.Value = messages.Messages.Timestamp(
					user.Locale,
					ifaceVal.(time.Time),
				)
			}
		case reflect.Bool:
			if ifaceVal.(bool) {
				item.Value = messages.Messages.Get(user.Locale, "yes")
			} else {
				item.Value = messages.Messages.Get(user.Locale, "no")
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

	item.Name = field.humanName(user.Locale)
	return item
}
