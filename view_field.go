package prago

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"time"
)

type viewField struct {
	Icon string
	Name string

	ViewContent *viewFieldContent

	Content    template.HTML
	EditAction template.JS
	EditName   string
}

type viewFieldContent struct {
	Name string
}

func (vf *viewFieldContent) IsEmpty() bool {
	if vf == nil {
		return true
	}
	if vf.Name != "" {
		return false
	}
	return true

}

func (resource *Resource) getBoxHeader(id int64, item any, request *Request) *boxHeader {
	ret := &boxHeader{}
	ret.DescriptionsBefore = []string{fmt.Sprintf("%s #%d", resource.singularName(request.Locale()), id)}
	ret.Name = resource.previewer(request, item).Name()
	ret.Icon = iconView
	ret.Image = resource.previewer(request, item).ImageURL()
	ret.Buttons = resource.getItemButtonData(request, item, true)
	return ret

}

func (resource *Resource) getViewFields(id int64, item any, request *Request) (ret []viewField) {

	for i, field := range resource.fields {
		if !field.authorizeView(request) {
			continue
		}

		if field.id == "id" {
			continue
		}

		var ifaceVal any
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(item).Elem().Field(i),
		)

		var editURL string
		if field.authorizeEdit(request) {
			editURL = resource.getURL(fmt.Sprintf("%d/edit?_focus=%s&_fields=%s", id, field.id, field.id))
		}

		var contentOLD template.HTML

		var viewContent *viewFieldContent
		if field.fieldType.getViewFieldContent != nil {
			viewContent = field.fieldType.getViewFieldContent(request, ifaceVal)
		} else {
			contentOLD = resource.app.adminTemplates.ExecuteToHTML(
				field.fieldType.viewTemplate,
				field.fieldType.viewDataSource(request, field, ifaceVal),
			)
		}

		kind := field.typ.Kind()
		if kind == reflect.Float64 || kind == reflect.Int64 || kind == reflect.Int {
			if contentOLD == "0" {
				contentOLD = ""
			}
		}

		contentOLD = template.HTML(strings.Trim(string(contentOLD), " \n\t"))

		if contentOLD == "" && viewContent.IsEmpty() {
			continue
		}

		icon := field.getIcon()

		vf := viewField{
			Icon:        icon,
			Name:        field.name(request.Locale()),
			ViewContent: viewContent,
			Content:     contentOLD,
			EditName:    fmt.Sprintf("Upravit položku „%s“", field.name(request.Locale())),
		}

		if editURL != "" {
			vf.EditAction = template.JS(fmt.Sprintf("popup(\"%s\")", editURL))
		}

		ret = append(
			ret,
			vf,
		)
	}

	for _, v := range resource.itemStats {
		if !request.Authorize(v.Permission) {
			continue
		}
		ret = append(
			ret,
			viewField{
				Name:    v.Name(request.Locale()),
				Content: template.HTML(v.Handler(item)),
			},
		)
	}

	return ret
}

func defaultStringer(userData UserData, field *Field, value any) string {
	return fmt.Sprintf("%v", value)
}

func numberStringer(userData UserData, field *Field, value any) string {
	return humanizeNumber(value.(int64))
}

func floatStringer(userData UserData, f *Field, value any) string {
	return humanizeFloat(value.(float64), userData.Locale())
}

func dateStringer(userData UserData, f *Field, value any) string {
	return messages.Timestamp(
		userData.Locale(),
		value.(time.Time),
		false,
	)
}

func timeStringer(userData UserData, field *Field, value any) string {
	return messages.Timestamp(
		userData.Locale(),
		value.(time.Time),
		true,
	)
}

func boolStringer(userData UserData, field *Field, value any) string {
	if value.(bool) {
		return messages.Get(userData.Locale(), "yes")
	}
	return ""
}

func stringerToDataSource(fn func(userData UserData, field *Field, value any) string) func(request *Request, field *Field, value any) any {
	return func(userData *Request, field *Field, value any) any {
		retStr := fn(userData, field, value)
		return any(retStr)
	}
}
