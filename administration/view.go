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
	URL  string
	Name string
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

func getRelationViewData(resource Resource, user User, f Field, value interface{}) interface{} {
	var val viewRelationData
	var relationName string
	if f.Tags["prago-relation"] != "" {
		relationName = f.Tags["prago-relation"]
	} else {
		relationName = f.Name
	}

	r2 := resource.Admin.getResourceByName(relationName)
	if r2 == nil {
		val.Name = fmt.Sprintf("Resource '%s' not found", relationName)
		return val
	}

	if !resource.Admin.Authorize(user, r2.CanView) {
		val.Name = fmt.Sprintf("User is not authorized to view this item")
		return val
	}

	var item interface{}
	r2.newItem(&item)
	err := resource.Admin.Query().WhereIs("id", value.(int64)).Get(item)
	if err != nil {
		val.Name = fmt.Sprintf("Can't find this item")
		return val
	}

	val.Name = getItemName(item, user.Locale)
	val.URL = r2.GetItemURL(item, "")
	return val
}
