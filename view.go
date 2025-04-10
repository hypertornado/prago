package prago

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
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
	Content  template.HTML
	EditURL  string
	EditName string
}

type viewButton struct {
	Name string
	URL  string
	Icon string
}

func (resource *Resource) getViews(item any, request *Request) (ret []*view) {
	id := resource.previewer(request, item).ID()
	ret = append(ret, resource.getBasicView(id, item, request))
	ret = append(ret, resource.getRelationViews(id, request)...)
	return ret
}

func (resource *Resource) getBasicView(id int64, item any, request *Request) *view {
	ret := &view{
		Header: &boxHeader{},
	}

	tableIcon := resource.icon
	if tableIcon == "" {
		tableIcon = iconTable
	}

	ret.Header.Name = resource.previewer(request, item).Name()
	ret.Header.Icon = iconView
	ret.Header.Image = resource.previewer(request, item).ImageURL()
	ret.Header.Buttons = resource.getItemButtonData(request, item, true)

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

		var content template.HTML
		if f.viewContentGenerator != nil {
			content = f.viewContentGenerator(ifaceVal)
		} else {
			content = resource.app.adminTemplates.ExecuteToHTML(
				f.fieldType.viewTemplate,
				f.fieldType.viewDataSource(request, f, ifaceVal),
			)
		}

		icon := f.getIcon()
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:     icon,
				Name:     f.name(request.Locale()),
				Content:  content,
				EditURL:  editURL,
				EditName: messages.Get(request.Locale(), "admin_edit"),
			},
		)
	}

	for _, v := range resource.itemStats {
		if !request.Authorize(v.Permission) {
			continue
		}
		ret.Items = append(
			ret.Items,
			viewField{
				Icon:    "glyphicons-basic-43-stats-circle.svg",
				Name:    v.Name(request.Locale()),
				Content: template.HTML(v.Handler(item)),
			},
		)
	}

	return ret
}

func getDefaultViewTemplate(_ reflect.Type) string {
	return "view_text"
}

func getDefaultViewDataSource(_ *Field) func(request *Request, f *Field, value interface{}) interface{} {
	return func(request *Request, f *Field, value interface{}) interface{} {
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
