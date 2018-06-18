package administration

import (
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

func (resource Resource) getView(inValues interface{}, user User) view {
	visible := defaultVisibilityFilter
	ret := view{}
	for i, f := range resource.fieldArrays {
		if !visible(resource, user, *f) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(inValues).Elem().Field(i),
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name:     f.HumanName(user.Locale),
				Template: f.fieldType.ViewTemplate,
				Value:    f.fieldType.ViewDataSource(resource, user, *f, ifaceVal),
			},
		)
	}
	return ret
}

func getDefaultViewTemplate(t reflect.Type) string {
	return "admin_item_view_text"
}

func getDefaultViewDataSource(t reflect.Type) func(resource Resource, user User, f Field, value interface{}) interface{} {
	if t == reflect.TypeOf(time.Now()) {
		return timestampViewDataSource
	}
	switch t.Kind() {
	case reflect.Bool:
		return boolViewDataSource
	default:
		return defaultViewDataSource
	}
}

func defaultViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return value
}

func timestampViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return messages.Messages.Timestamp(
		user.Locale,
		value.(time.Time),
	)
}

func boolViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	if value.(bool) {
		return messages.Messages.Get(user.Locale, "yes")
	} else {
		return messages.Messages.Get(user.Locale, "no")
	}
}
