package prago

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Preview struct {
	ID          int64
	Image       string
	ImageID     string
	URL         string
	Name        string
	Description string
}

type previewer struct {
	userData UserData
	item     any
	resource *Resource
}

func (resource *Resource) previewer(userData UserData, item any) *previewer {
	if !resource.isItPointerToResourceItem(item) {
		return nil
	}

	return &previewer{
		userData: userData,
		item:     item,
		resource: resource,
	}
}

func (previewer *previewer) hasAccessToField(fieldID string) bool {
	if !previewer.userData.Authorize(previewer.resource.canView) {
		return false
	}

	fieldID = strings.ToLower(fieldID)
	field := previewer.resource.fieldMap[fieldID]
	if field == nil {
		return false
	}
	return field.authorizeView(previewer.userData)
}

func (previewer *previewer) ID() int64 {
	if previewer.item == nil {
		return -1
	}
	itemsVal := reflect.ValueOf(previewer.item).Elem()
	field := itemsVal.FieldByName("ID")
	if field.IsValid() {
		return field.Int()
	}
	return -1
}

type namedIFace interface {
	GetName(string) string
}

func (previewer *previewer) Name() string {
	pointerVal := reflect.ValueOf(previewer.item)
	itemsVal := pointerVal.Elem()
	var valIface = pointerVal.Interface()
	namedIface, ok := valIface.(namedIFace)
	if ok {
		custom := namedIface.GetName(previewer.userData.Locale())
		if custom != "" {
			return custom
		}
	}

	if previewer.item != nil && previewer.hasAccessToField("Name") {
		field := itemsVal.FieldByName("Name")
		if field.IsValid() {
			ret := field.String()
			if ret != "" {
				return ret
			}
		}
	}
	return fmt.Sprintf("#%d", previewer.ID())

}

func (f *Field) relationPreview(userData UserData, idsStr string) (ret []*Preview) {
	ids := strings.Split(idsStr, ";")
	for _, id := range ids {
		item := f.relatedResource.query(context.Background()).ID(id)
		if item == nil {
			continue
		}
		ret = append(ret, f.relatedResource.previewer(userData, item).Preview(f.resource))
	}

	return
}

func (previewer *previewer) URL(suffix string) string {
	return previewer.resource.getItemURL(previewer.item, suffix, previewer.userData)
}

func (previewer *previewer) Preview(relatedResource *Resource) *Preview {
	var ret Preview
	ret.ID = previewer.ID()
	ret.Name = previewer.Name()
	ret.URL = previewer.URL("")
	ret.Image = previewer.ThumbnailURL()
	ret.ImageID = previewer.ThumbnailID()
	ret.Description = previewer.DescriptionExtended(relatedResource)
	return &ret
}

func (previewer *previewer) ThumbnailID() string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return field.String()
		}
	}
	return ""
}

func (previewer *previewer) ThumbnailURL() string {
	id := previewer.ThumbnailID()
	if id != "" {
		return previewer.resource.app.thumb(id)
	}
	return ""
}

func (previewer *previewer) ImageURL() string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return previewer.resource.app.largeImage(field.String())
		}
	}
	return ""
}

func (previewer *previewer) DescriptionBasic(relatedResource *Resource) string {
	itemsVal := reflect.ValueOf(previewer.item).Elem()

	if previewer.item != nil {
		field := itemsVal.FieldByName("Description")
		if field.IsValid() && previewer.hasAccessToField("Description") {
			ret := field.String()
			croped := cropMarkdown(ret, 200)
			if croped != "" {
				return croped
			}
		}
	}
	return ""
}

func (previewer *previewer) DescriptionExtended(relatedResource *Resource) string {
	var items []string
	itemsVal := reflect.ValueOf(previewer.item).Elem()

	basicDescription := previewer.DescriptionBasic(relatedResource)

	if basicDescription != "" {
		items = append(items, basicDescription)
	}

	for _, v := range previewer.resource.fields {
		if v.fieldClassName == "ID" || v.fieldClassName == "Name" || v.fieldClassName == "Description" {
			continue
		}
		if !v.authorizeView(previewer.userData) {
			continue
		}

		rr := v.relatedResource
		if rr != nil && relatedResource != nil && rr.getID() == relatedResource.getID() {
			continue
		}

		field := itemsVal.FieldByName(v.fieldClassName)
		stringed := previewer.resource.app.relationStringer(*v, field, previewer.userData)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.name(previewer.userData.Locale()), stringed))
		}
	}
	ret := strings.Join(items, " · ")
	return ret
}
