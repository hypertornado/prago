package prago

import (
	"fmt"
	"reflect"
	"time"
)

type view struct {
	Name       string
	Subname    string
	Navigation []tab
	Items      []viewField
	Relation   *viewRelation
}

type viewField struct {
	Name     string
	Template string
	Value    interface{}
}

func (resource *Resource[T]) getViews(id int, inValues interface{}, user *user) (ret []view) {
	ret = append(ret, resource.getBasicView(id, inValues, user))
	ret = append(ret, resource.getAutoRelationsView(id, inValues, user)...)
	return ret
}

func (resource *Resource[T]) getBasicView(id int, inValues interface{}, user *user) view {
	ret := view{}

	ret.Items = append(
		ret.Items,
		viewField{
			Name:     messages.Get(user.Locale, "admin_table"),
			Template: "admin_item_view_url",
			Value: [2]string{
				resource.getURL(""),
				resource.name(user.Locale),
			},
		},
	)

	for i, f := range resource.fields {
		if !f.authorizeView(user) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(inValues).Elem().Field(i),
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name:     f.humanName(user.Locale),
				Template: f.fieldType.viewTemplate,
				Value:    f.fieldType.viewDataSource(user, f, ifaceVal),
			},
		)
	}

	historyView := resource.app.getHistory(resource, int64(id))

	if len(historyView.Items) > 0 {
		ret.Items = append(
			ret.Items,
			viewField{
				Name:     messages.Get(user.Locale, "admin_history_last"),
				Template: "admin_item_view_url",
				Value: [2]string{
					historyView.Items[0].UserURL,
					historyView.Items[0].UserName,
				},
			},
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name:     messages.Get(user.Locale, "admin_history_count"),
				Template: "admin_item_view_url",
				Value: [2]string{
					resource.getURL(fmt.Sprintf("%d/history", id)),
					fmt.Sprintf("%d", len(historyView.Items)),
				},
			},
		)

	}

	return ret
}

func getDefaultViewTemplate(t reflect.Type) string {
	return "admin_item_view_text"
}

func getDefaultViewDataSource(f *Field) func(user *user, f *Field, value interface{}) interface{} {
	t := f.typ
	if t == reflect.TypeOf(time.Now()) {
		if f.tags["prago-type"] == "timestamp" || f.fieldClassName == "CreatedAt" || f.fieldClassName == "UpdatedAt" {
			return timestampViewDataSource
		}
		return timeViewDataSource
	}
	switch t.Kind() {
	case reflect.Bool:
		return boolViewDataSource
	case reflect.Int:
		return numberViewDataSource
	case reflect.Int64:
		return numberViewDataSource
	case reflect.Float64:
		return floatViewDataSource
	default:
		return defaultViewDataSource
	}
}

func defaultViewDataSource(user *user, f *Field, value interface{}) interface{} {
	return value
}

func numberViewDataSource(user *user, f *Field, value interface{}) interface{} {
	switch f.typ.Kind() {
	case reflect.Int:
		return humanizeNumber(int64(value.(int)))
	case reflect.Int64:
		return humanizeNumber(value.(int64))
	}

	return value
}

func floatViewDataSource(user *user, f *Field, value interface{}) interface{} {
	return humanizeFloat(value.(float64), user.Locale)
}

func timeViewDataSource(user *user, f *Field, value interface{}) interface{} {
	return messages.Timestamp(
		user.Locale,
		value.(time.Time),
		false,
	)
}

func timestampViewDataSource(user *user, f *Field, value interface{}) interface{} {
	return messages.Timestamp(
		user.Locale,
		value.(time.Time),
		true,
	)
}

func boolViewDataSource(user *user, f *Field, value interface{}) interface{} {
	if value.(bool) {
		return messages.Get(user.Locale, "yes")
	}
	return messages.Get(user.Locale, "no")
}
