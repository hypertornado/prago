package prago

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
)

type view struct {
	Name       string
	Subname    string
	Navigation []navigationTab
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

func (resource Resource) getViews(id int, inValues interface{}, user User) (ret []view) {
	ret = append(ret, resource.getBasicView(id, inValues, user))
	ret = append(ret, resource.getAutoRelationsView(id, inValues, user)...)
	return ret
}

func (resource Resource) getBasicView(id int, inValues interface{}, user User) view {
	visible := defaultVisibilityFilter
	ret := view{}

	ret.Items = append(
		ret.Items,
		viewField{
			Name:     messages.Messages.Get(user.Locale, "admin_table"),
			Template: "admin_item_view_url",
			Value: [2]string{
				resource.GetURL(""),
				resource.HumanName(user.Locale),
			},
		},
	)

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

	historyView := resource.App.getHistory(&resource, int64(id))

	if len(historyView.Items) > 0 {
		ret.Items = append(
			ret.Items,
			viewField{
				Name:     messages.Messages.Get(user.Locale, "admin_history_last"),
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
				Name:     messages.Messages.Get(user.Locale, "admin_history_count"),
				Template: "admin_item_view_url",
				Value: [2]string{
					resource.GetURL(fmt.Sprintf("%d/history", id)),
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

func getDefaultViewDataSource(f *Field) func(resource Resource, user User, f Field, value interface{}) interface{} {
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

func defaultViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return value
}

func numberViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	switch f.Typ.Kind() {
	case reflect.Int:
		return utils.HumanizeNumber(int64(value.(int)))
	case reflect.Int64:
		return utils.HumanizeNumber(value.(int64))
	}

	return value
}

func floatViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return utils.HumanizeFloat(value.(float64), user.Locale)
}

func timeViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return messages.Messages.Timestamp(
		user.Locale,
		value.(time.Time),
		false,
	)
}

func timestampViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return messages.Messages.Timestamp(
		user.Locale,
		value.(time.Time),
		true,
	)
}

func boolViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	if value.(bool) {
		return messages.Messages.Get(user.Locale, "yes")
	}
	return messages.Messages.Get(user.Locale, "no")
}
