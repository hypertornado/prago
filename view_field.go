package prago

import (
	"fmt"
	"html/template"
	"reflect"
)

type viewField struct {
	Icon     string
	Name     string
	SubNames []string

	ViewContent *viewFieldContent

	Buttons []*Button

	Relation *viewRelationItem
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
	previewer := resource.previewer(request, item)
	ret.Name = previewer.Name()
	ret.Icon = previewer.Icon()
	ret.Image = previewer.ImageURL()

	style := previewer.Style()
	if style != "" {
		ret.Style = style
	}

	ret.Buttons = resource.getItemButtonData(request, item, true)
	return ret

}

func (resource *Resource) getViewFields(id int64, item any, request *Request) (ret []*viewField) {

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

		vf := &viewField{
			Icon:        icon,
			Name:        field.name(request.Locale()),
			ViewContent: viewContent,
		}

		if editURL != "" {
			vf.Buttons = append(vf.Buttons, &Button{
				Name:    fmt.Sprintf("Upravit položku „%s“", field.name(request.Locale())),
				Icon:    "glyphicons-basic-31-pencil.svg",
				OnClick: template.JS(fmt.Sprintf("popup(\"%s\")", editURL)),
			})
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
			&viewField{
				Name: v.Name(request.Locale()),
				ViewContent: &viewFieldContent{
					ContentHTML: template.HTML(v.Handler(item)),
				},
			},
		)
	}

	return ret
}

func defaultViewFieldContent(request *Request, field *Field, val any) *viewFieldContent {
	name := fmt.Sprintf("%v", val)
	var empty bool
	if name == "" {
		empty = true
	}
	return &viewFieldContent{
		Empty: empty,
		Name:  name,
	}

}
