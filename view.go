package prago

import (
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

type view struct {
	Icon       string
	Name       string
	Subname    string
	Navigation []viewButton
	Header     *boxHeader
	Items      []viewField
	Relation   *viewRelation

	SearchResults []*searchItem
	Pagination    []paginationItem
}

type viewField struct {
	Icon     string
	Name     string
	Template string
	Value    interface{}
	EditURL  string
}

type viewButton struct {
	URL  string
	Icon string
}

func (resource *Resource) getViews(ctx context.Context, item any, request *Request) (ret []*view) {
	id := resource.previewer(request, item).ID()
	ret = append(ret, resource.getBasicView(ctx, id, item, request))
	ret = append(ret, resource.getRelationViews(ctx, id, request)...)
	return ret
}

func (resource *Resource) getBasicView(ctx context.Context, id int64, item any, request *Request) *view {
	ret := &view{
		Header: &boxHeader{},
	}

	tableIcon := resource.icon
	if tableIcon == "" {
		tableIcon = iconTable
	}

	/*
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:     tableIcon,
				Name:     messages.Get(request.Locale(), "admin_table"),
				Template: "admin_item_view_url",
				Value: [2]string{
					resource.getURL(""),
					resource.pluralName(request.Locale()),
				},
			},
		)*/

	ret.Header.Name = resource.previewer(request, item).Name()
	ret.Header.Icon = iconView
	ret.Header.Image = resource.previewer(request, item).ImageURL(ctx)
	ret.Header.Buttons = resource.getItemButtonData(request, item)

	resourceIcon := resource.icon
	if resourceIcon == "" {
		resourceIcon = iconResource
	}

	for i, f := range resource.fields {
		if !f.authorizeView(request) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(item).Elem().Field(i),
		)

		var editURL string
		if f.authorizeEdit(request) {
			editURL = resource.getURL(fmt.Sprintf("%d/edit?_focus=%s", id, f.id))
		}

		icon := f.getIcon()
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:     icon,
				Name:     f.name(request.Locale()),
				Template: f.fieldType.viewTemplate,
				Value:    f.fieldType.viewDataSource(ctx, request, f, ifaceVal),
				EditURL:  editURL,
			},
		)
	}

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
