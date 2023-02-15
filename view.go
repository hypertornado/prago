package prago

import (
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

type view struct {
	Icon         string
	Name         string
	Subname      string
	Navigation   []viewButton
	Header       *BoxHeader
	Items        []viewField
	Relation     *viewRelation
	QuickActions []QuickActionView
}

type viewField struct {
	Icon     string
	Name     string
	Template string
	Value    interface{}
}

type viewButton struct {
	URL  string
	Icon string
}

func (resourceData *resourceData) getViews(ctx context.Context, item any, request *Request) (ret []view) {
	id := resourceData.previewer(request, item).ID()
	ret = append(ret, resourceData.getBasicView(ctx, id, item, request))
	ret = append(ret, resourceData.getRelationViews(ctx, id, request)...)
	return ret
}

func (resourceData *resourceData) getBasicView(ctx context.Context, int64, item any, request *Request) view {
	ret := view{
		QuickActions: resourceData.getQuickActionViews(item, request),
		Header:       &BoxHeader{},
	}

	tableIcon := resourceData.icon
	if tableIcon == "" {
		tableIcon = iconTable
	}

	ret.Items = append(
		ret.Items,
		viewField{
			Icon:     tableIcon,
			Name:     messages.Get(request.Locale(), "admin_table"),
			Template: "admin_item_view_url",
			Value: [2]string{
				resourceData.getURL(""),
				resourceData.pluralName(request.Locale()),
			},
		},
	)

	ret.Header.Name = resourceData.previewer(request, item).Name()
	ret.Header.Icon = tableIcon
	ret.Header.Image = resourceData.previewer(request, item).ImageURL(ctx)

	resourceIcon := resourceData.icon
	if resourceIcon == "" {
		resourceIcon = iconResource
	}
	ret.Header.Tags = append(ret.Header.Tags, BoxTag{
		URL:  fmt.Sprintf("/admin/%s", resourceData.id),
		Icon: resourceIcon,
		Name: resourceData.pluralName(request.Locale()),
	})

	for i, f := range resourceData.fields {
		if !f.authorizeView(request) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(item).Elem().Field(i),
		)

		icon := f.getIcon()
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:     icon,
				Name:     f.name(request.Locale()),
				Template: f.fieldType.viewTemplate,
				Value:    f.fieldType.viewDataSource(ctx, request, f, ifaceVal),
			},
		)
	}

	/*historyView := resourceData.app.getHistory(resourceData, int64(id))

	if len(historyView.Items) > 0 {
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:     "glyphicons-basic-58-history.svg",
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
				Icon:     "glyphicons-basic-58-history.svg",
				Name:     messages.Get(user.Locale, "admin_history_count"),
				Template: "admin_item_view_url",
				Value: [2]string{
					resourceData.getURL(fmt.Sprintf("%d/history", id)),
					fmt.Sprintf("%d", len(historyView.Items)),
				},
			},
		)

	}*/

	return ret
}

func getDefaultViewTemplate(t reflect.Type) string {
	return "admin_item_view_text"
}

func getDefaultViewDataSource(f *Field) func(ctx context.Context, request *Request, f *Field, value interface{}) interface{} {
	return func(ctx context.Context, request *Request, f *Field, value interface{}) interface{} {
		return getDefaultFieldStringer(f)(request, f, value)
	}
}

func getDefaultFieldStringer(f *Field) func(userData UserData, f *Field, value interface{}) string {
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

func defaultViewDataSource(userData UserData, f *Field, value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func numberViewDataSource(userData UserData, f *Field, value interface{}) string {
	switch f.typ.Kind() {
	case reflect.Int:
		return humanizeNumber(int64(value.(int)))
	case reflect.Int64:
		return humanizeNumber(value.(int64))
	}
	panic("not integer type")
	//return defaultViewDataSource(user, f, value)
}

func floatViewDataSource(userData UserData, f *Field, value interface{}) string {
	return humanizeFloat(value.(float64), userData.Locale())
}

func timeViewDataSource(userData UserData, f *Field, value interface{}) string {
	return messages.Timestamp(
		userData.Locale(),
		value.(time.Time),
		false,
	)
}

func timestampViewDataSource(userData UserData, f *Field, value interface{}) string {
	return messages.Timestamp(
		userData.Locale(),
		value.(time.Time),
		true,
	)
}

func boolViewDataSource(userData UserData, f *Field, value interface{}) string {
	if value.(bool) {
		return messages.Get(userData.Locale(), "yes")
	}
	return messages.Get(userData.Locale(), "no")
}
