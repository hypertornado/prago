package administration

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hypertornado/prago/administration/messages"
)

type view struct {
	Items []viewField
}

type viewField struct {
	Name     string
	Button   *viewFieldAction
	Template string
	Value    interface{}
}

type viewFieldAction struct {
	Name string
	URL  string
}

func (resource Resource) getView(id int, inValues interface{}, user User) view {
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

	historyView := resource.Admin.getHistory(&resource, int64(id))

	if len(historyView.Items) > 0 {
		ret.Items = append(
			ret.Items,
			viewField{
				Name:     messages.Messages.Get(user.Locale, "admin_history_last"),
				Template: "admin_item_view_text",
				Value:    historyView.Items[0].UserName,
			},
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name: messages.Messages.Get(user.Locale, "admin_history_count"),
				Button: &viewFieldAction{
					Name: messages.Messages.Get(user.Locale, "admin_history"),
					URL:  resource.GetURL(fmt.Sprintf("%d/history", id)),
				},
				Template: "admin_item_view_text",
				Value:    fmt.Sprintf("%d", len(historyView.Items)),
			},
		)

	}

	for _, v := range resource.relations {
		q := resource.Admin.prefilterQuery(v.field, fmt.Sprintf("%d", id))
		if v.resource.OrderDesc {
			q = q.OrderDesc(v.resource.OrderByColumn)
		} else {
			q = q.Order(v.resource.OrderByColumn)
		}

		q = q.Limit(resource.ItemsPerPage)

		var rowItems interface{}

		v.resource.newArrayOfItems(&rowItems)
		q.Get(rowItems)

		vv := reflect.ValueOf(rowItems).Elem()
		var data []interface{}
		for i := 0; i < vv.Len(); i++ {
			data = append(
				data,
				v.resource.itemToRelationData(vv.Index(i).Interface()),
			)
		}

		addURL := resource.GetURL(fmt.Sprintf("%d/%s", id, v.addURL()))

		name := v.resource.HumanName(user.Locale)
		ret.Items = append(
			ret.Items,
			viewField{
				Name: name,
				Button: &viewFieldAction{
					Name: v.addName(user.Locale),
					URL:  addURL,
				},
				Template: "admin_item_view_relations",
				Value:    data,
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
