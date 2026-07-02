package prago

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
)

type viewField struct {
	Icon string
	Name string

	ViewContent *viewFieldContent

	//Content    template.HTML
	EditAction template.JS
	EditName   string
}

type viewFieldContent struct {
	Empty bool

	Icon string
	Name string

	Color string
	Style string

	ContentHTML template.HTML

	Previews    []*Preview
	CDNFileData *cdnFileData

	Images *ImagePickerResponse

	PlaceData string

	VideoURL string
}

func (vfc *viewFieldContent) HasNameOrIcon() bool {
	if vfc.Name != "" {
		return true
	}
	if vfc.Icon != "" {
		return true
	}
	return false
}

func (vfc *viewFieldContent) IconColor() string {
	if vfc.Color != "" {
		return vfc.Color
	}
	return getStyleColor(vfc.Style)
}

func (vf *viewFieldContent) IsEmpty() bool {
	if vf == nil {
		return true
	}
	if vf.Empty {
		return true
	}
	return false
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

		var viewContent *viewFieldContent
		if field.fieldType.getViewFieldContent != nil {
			viewContent = field.fieldType.getViewFieldContent(request, field, ifaceVal)
		} else {

		}

		if viewContent.IsEmpty() {
			continue
		}

		icon := field.getIcon()

		vf := viewField{
			Icon:        icon,
			Name:        field.name(request.Locale()),
			ViewContent: viewContent,
			//Content:     contentOLD,
			EditName: fmt.Sprintf("Upravit položku „%s“", field.name(request.Locale())),
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
				Name: v.Name(request.Locale()),
				ViewContent: &viewFieldContent{
					ContentHTML: template.HTML(v.Handler(item)),
				},
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

func stringerToViewFieldContent(fn func(userData UserData, field *Field, value any) string) func(request *Request, field *Field, val any) *viewFieldContent {
	return func(request *Request, field *Field, val any) *viewFieldContent {
		name := fn(request, field, val)
		var empty bool
		if name == "" {
			empty = true
		}
		return &viewFieldContent{
			Empty: empty,
			Name:  name,
		}
	}
}
