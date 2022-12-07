package prago

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type preview struct {
	ID          int64
	Image       string
	URL         string
	Name        string
	Description string
}

type previewer struct {
	user         *user
	item         any
	resourceData *resourceData
}

func (resourceData *resourceData) previewer(user *user, item any) *previewer {
	if reflect.PointerTo(resourceData.typ) != reflect.TypeOf(item) {
		return nil
	}

	/*if !resourceData.app.authorize(user, resourceData.canView) {
		return nil
	}*/

	return &previewer{
		user:         user,
		item:         item,
		resourceData: resourceData,
	}
}

func (previewer *previewer) hasAccessToField(fieldID string) bool {
	if !previewer.resourceData.app.authorize(previewer.user, previewer.resourceData.canView) {
		return false
	}

	fieldID = strings.ToLower(fieldID)
	field := previewer.resourceData.fieldMap[fieldID]
	if field == nil {
		return false
	}
	return field.authorizeView(previewer.user)
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
	if previewer.item != nil && previewer.hasAccessToField("Name") {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		var valIface = itemsVal.Interface()
		namedIface, ok := valIface.(namedIFace)
		if ok {
			return namedIface.GetName(previewer.user.Locale)
		}
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

func (f *Field) relationPreview(ctx context.Context, user *user, id int64) *preview {
	item := f.relatedResource.query(ctx).ID(id)
	if item == nil {
		return nil
	}
	return f.relatedResource.previewer(user, item).Preview(ctx, f.resource)

}

func (previewer *previewer) URL(suffix string) string {
	return previewer.resourceData.getItemURL(previewer.item, suffix, previewer.user)
}

func (previewer *previewer) Preview(ctx context.Context, relatedResource *resourceData) *preview {
	var ret preview
	ret.ID = previewer.ID()
	ret.Name = previewer.Name()
	ret.URL = previewer.URL("")
	ret.Image = previewer.ThumbnailURL(ctx)
	ret.Description = previewer.DescriptionExtended(relatedResource)
	return &ret
}

func (previewer *previewer) ThumbnailURL(ctx context.Context) string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return previewer.resourceData.app.thumb(ctx, field.String())
		}
	}
	return ""
}

func (previewer *previewer) ImageURL(ctx context.Context) string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return previewer.resourceData.app.largeImage(ctx, field.String())
		}
	}
	return ""
}

func (previewer *previewer) DescriptionBasic(relatedResource *resourceData) string {
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

func (previewer *previewer) DescriptionExtended(relatedResource *resourceData) string {
	var items []string
	itemsVal := reflect.ValueOf(previewer.item).Elem()

	basicDescription := previewer.DescriptionBasic(relatedResource)

	if basicDescription != "" {
		items = append(items, basicDescription)
	}

	for _, v := range previewer.resourceData.fields {
		if v.fieldClassName == "ID" || v.fieldClassName == "Name" || v.fieldClassName == "Description" {
			continue
		}
		if !v.authorizeView(previewer.user) {
			continue
		}

		rr := v.relatedResource
		if rr != nil && relatedResource != nil && rr.getID() == relatedResource.getID() {
			continue
		}

		field := itemsVal.FieldByName(v.fieldClassName)
		stringed := previewer.resourceData.app.relationStringer(*v, field, previewer.user)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.name(previewer.user.Locale), stringed))
		}
	}
	ret := strings.Join(items, " · ")
	return cropMarkdown(ret, 500)
}
