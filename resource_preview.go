package prago

import (
	"errors"
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
	prev, err := resourceData.getPreviewer(user, item)
	must(err)
	return prev
}

func (resourceData *resourceData) getPreviewer(user *user, item any) (*previewer, error) {
	if reflect.PointerTo(resourceData.typ) != reflect.TypeOf(item) {
		return nil, errors.New("wrong type of previewer item")
	}

	if !resourceData.app.authorize(user, resourceData.canView) {
		return nil, errors.New("can't view this item")
	}

	return &previewer{
		user:         user,
		item:         item,
		resourceData: resourceData,
	}, nil
}

func (previewer *previewer) hasAccessToField(fieldID string) bool {
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
	if field.IsValid() && previewer.hasAccessToField("ID") {
		return field.Int()
	}
	return -1

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

/*func (resourceData *resourceData) getItemName(item interface{}) string {
	//TODO: Authorize field
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		var valIface = itemsVal.Interface()
		namedIface, ok := valIface.(namedIFace)
		if ok {
			return namedIface.GetName()
		}
		field := itemsVal.FieldByName("Name")
		if field.IsValid() {
			ret := field.String()
			if ret != "" {
				return ret
			}
		}
	}
	return fmt.Sprintf("#%d", resourceData.getItemID(item))
}*/

/*func (resourceData *resourceData) resourceItemName(user *user, id int64) string {
	item := resourceData.query().ID(id)
	if item == nil {
		return fmt.Sprintf("%d - not found", id)
	}
	return resourceData.previewer(user, item).Name()
}*/

/*func (resourceData *resourceData) getItemID(item interface{}) int64 {
	if item == nil {
		return -1
	}

	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("ID")
	if field.IsValid() {
		return field.Int()
	}
	return -1
}*/

func getRelationViewData(user *user, f *Field, value interface{}) interface{} {
	ret, _ := getPreviewData(user, f, value.(int64))
	return ret
}

func getPreviewData(user *user, f *Field, value int64) (*preview, error) {
	app := f.resource.app
	if f.relatedResource == nil {
		return nil, fmt.Errorf("resource not found: %s", f.name("en"))
	}

	if !app.authorize(user, f.relatedResource.canView) {
		return nil, fmt.Errorf("user is not authorized to view this item")
	}

	item := f.relatedResource.query().ID(value)
	if item == nil {
		return nil, errors.New("can't get item preview")
	}

	ret := f.relatedResource.getPreview(item, user, nil)
	if ret == nil {
		return nil, errors.New("can't get item preview")
	}

	return ret, nil
}

/*func (resourceData *resourceData) getItemPreview(id int64, user *user, relatedResource *resourceData) *preview {
	item := resourceData.query().ID(id)
	if item == nil {
		return nil
	}
	return resourceData.getPreview(item, user, nil)
}*/

func (resourceData *resourceData) getPreview(item any, user *user, relatedResource *resourceData) *preview {
	var ret preview
	ret.ID = resourceData.previewer(user, item).ID()
	ret.Name = resourceData.previewer(user, item).Name()
	ret.URL = resourceData.getItemURL(item, "", user)
	ret.Image = resourceData.previewer(user, item).ThumbnailURL()
	ret.Description = resourceData.getItemDescription(item, user, relatedResource)
	return &ret
}

func (previewer *previewer) ThumbnailURL() string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return previewer.resourceData.app.thumb(field.String())
		}
	}
	return ""
}

func (previewer *previewer) ImageURL() string {
	if previewer.item != nil {
		itemsVal := reflect.ValueOf(previewer.item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() && previewer.hasAccessToField("Image") {
			return previewer.resourceData.app.largeImage(field.String())
		}
	}
	return ""
}

type namedIFace interface {
	GetName(string) string
}

func (resourceData *resourceData) getItemDescription(item any, user *user, relatedResource *resourceData) string {
	var items []string
	itemsVal := reflect.ValueOf(item).Elem()

	//TODO: check access

	if item != nil {
		field := itemsVal.FieldByName("Description")
		if field.IsValid() {
			ret := field.String()
			croped := cropMarkdown(ret, 200)
			if croped != "" {
				items = append(items, croped)
			}
		}
	}

	for _, v := range resourceData.fields {
		if v.fieldClassName == "ID" || v.fieldClassName == "Name" || v.fieldClassName == "Description" {
			continue
		}
		if !v.authorizeView(user) {
			continue
		}

		rr := v.relatedResource
		if rr != nil && relatedResource != nil && rr.getID() == relatedResource.getID() {
			continue
		}

		field := itemsVal.FieldByName(v.fieldClassName)
		stringed := resourceData.app.relationStringer(*v, field, user)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.name(user.Locale), stringed))
		}
	}
	ret := strings.Join(items, " · ")
	return cropMarkdown(ret, 500)
}
