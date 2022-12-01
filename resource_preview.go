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

/*func (resourceData *resourceData) previewerFromID(user *user, id int64) (*previewer, error) {
	item := resourceData.query().ID(id)
	if item == nil {
		return nil, errors.New("can't find item for previewer")
	}
	return resourceData.getPreviewer(user, item)
}*/

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

/*func getRelationViewData(user *user, f *Field, value interface{}) interface{} {
	return f.relationPreview(user, value.(int64))
}*/

func (f *Field) relationPreview(user *user, id int64) *preview {
	item := f.relatedResource.query().ID(id)
	if item == nil {
		return nil
	}
	return f.relatedResource.previewer(user, item).Preview(f.resource)

	/*previewer, err := f.relatedResource.previewerFromID(user, id)
	if err != nil {
		return nil, err
	}

	return previewer.Preview(f.resource), nil*/

	/*app := f.resource.app
	if f.relatedResource == nil {
		return nil, fmt.Errorf("resource not found: %s", f.name("en"))
	}

	if !app.authorize(user, f.relatedResource.canView) {
		return nil, fmt.Errorf("user is not authorized to view this item")
	}

	item := f.relatedResource.query().ID(id)
	if item == nil {
		return nil, errors.New("can't get item preview")
	}

	ret := f.relatedResource.previewer(user, item).Preview(f.resource)
	if ret == nil {
		return nil, errors.New("can't get item preview")
	}

	return ret, nil*/
}

/*func (resourceData *resourceData) getItemPreview(id int64, user *user, relatedResource *resourceData) *preview {
	item := resourceData.query().ID(id)
	if item == nil {
		return nil
	}
	return resourceData.getPreview(item, user, nil)
}*/

func (previewer *previewer) URL(suffix string) string {
	return previewer.resourceData.getItemURL(previewer.item, suffix, previewer.user)
}

func (previewer *previewer) Preview(relatedResource *resourceData) *preview {
	var ret preview
	ret.ID = previewer.ID()
	ret.Name = previewer.Name()
	ret.URL = previewer.URL("")
	ret.Image = previewer.ThumbnailURL()
	ret.Description = previewer.Description(relatedResource)
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

func (previewer *previewer) Description(relatedResource *resourceData) string {
	var items []string
	itemsVal := reflect.ValueOf(previewer.item).Elem()

	if previewer.item != nil {
		field := itemsVal.FieldByName("Description")
		if field.IsValid() && previewer.hasAccessToField("Description") {
			ret := field.String()
			croped := cropMarkdown(ret, 200)
			if croped != "" {
				items = append(items, croped)
			}
		}
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
