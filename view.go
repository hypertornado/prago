package prago

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hypertornado/prago/utils"
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

type viewAction struct {
	Name string
	URL  string
}

func (resource Resource) getViews(id int, inValues interface{}, user *user) (ret []view) {
	ret = append(ret, resource.getBasicView(id, inValues, user))
	ret = append(ret, resource.getAutoRelationsView(id, inValues, user)...)
	return ret
}

func (resource Resource) getBasicView(id int, inValues interface{}, user *user) view {
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

	for i, f := range resource.fieldArrays {
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
				Name:     f.HumanName(user.Locale),
				Template: f.fieldType.viewTemplate,
				Value:    f.fieldType.viewDataSource(user, *f, ifaceVal),
			},
		)
	}

	historyView := resource.app.getHistory(&resource, int64(id))

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

func getDefaultViewDataSource(f *field) func(user *user, f field, value interface{}) interface{} {
	t := f.Typ
	if t == reflect.TypeOf(time.Now()) {
		if f.Tags["prago-type"] == "timestamp" || f.Name == "CreatedAt" || f.Name == "UpdatedAt" {
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

func defaultViewDataSource(user *user, f field, value interface{}) interface{} {
	return value
}

func numberViewDataSource(user *user, f field, value interface{}) interface{} {
	switch f.Typ.Kind() {
	case reflect.Int:
		return utils.HumanizeNumber(int64(value.(int)))
	case reflect.Int64:
		return utils.HumanizeNumber(value.(int64))
	}

	return value
}

func floatViewDataSource(user *user, f field, value interface{}) interface{} {
	return utils.HumanizeFloat(value.(float64), user.Locale)
}

func timeViewDataSource(user *user, f field, value interface{}) interface{} {
	return messages.Timestamp(
		user.Locale,
		value.(time.Time),
		false,
	)
}

func timestampViewDataSource(user *user, f field, value interface{}) interface{} {
	return messages.Timestamp(
		user.Locale,
		value.(time.Time),
		true,
	)
}

func boolViewDataSource(user *user, f field, value interface{}) interface{} {
	if value.(bool) {
		return messages.Get(user.Locale, "yes")
	}
	return messages.Get(user.Locale, "no")
}
