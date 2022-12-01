package prago

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type relatedField Field

type preview struct {
	ID          int64
	Image       string
	URL         string
	Name        string
	Description string
}

func (app *App) initRelations() {
	for _, resourceData := range app.resources {
		resourceData.createRelations()
	}
}

func (resourceData *resourceData) createRelations() {
	for _, field := range resourceData.fields {
		if field.tags["prago-type"] == "relation" {
			relatedResourceID := field.id
			if field.tags["prago-relation"] != "" {
				relatedResourceID = field.tags["prago-relation"]
			}
			field.relatedResource = resourceData.app.getResourceByID(relatedResourceID)

			if !field.nameSetManually {
				field.name = field.relatedResource.singularName
			}
			field.relatedResource.addRelation((*relatedField)(field))
		}
	}
}

func (field *relatedField) addURL(id int64) string {
	values := url.Values{}
	values.Add(field.id, fmt.Sprintf("%d", id))
	return field.resource.getURL("new?" + values.Encode())
}

func (field *relatedField) listURL(id int64) string {
	values := url.Values{}
	values.Add(field.id, fmt.Sprintf("%d", id))
	return field.resource.getURL("") + "?" + values.Encode()
}

func (field *relatedField) listName(locale string) string {
	f := (*Field)(field)
	ret := f.GetManuallySetPluralName(locale)
	if ret != "" {
		return fmt.Sprintf("%s – %s", field.resource.pluralName(locale), ret)
	}
	return field.resource.pluralName(locale)
}

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

	ret := f.relatedResource.getItemPreview(value, user, f.resource)
	if ret == nil {
		return nil, errors.New("can't get item preview")
	}
	return ret, nil
}

func (resourceData *resourceData) getItemPreview(id int64, user *user, relatedResource *resourceData) *preview {
	item := resourceData.query().ID(id)
	if item == nil {
		return nil
	}
	return resourceData.getPreview(item, user, nil)

}

func (resourceData *resourceData) getPreview(item any, user *user, relatedResource *resourceData) *preview {
	var ret preview
	ret.ID = getItemID(item)
	ret.Name = getItemName(item)
	ret.URL = resourceData.getItemURL(item, "")
	ret.Image = resourceData.app.getItemImage(item)
	ret.Description = resourceData.getItemDescription(item, user, relatedResource)
	return &ret
}

func (app *App) getItemImage(item interface{}) string {
	//TODO: Authorize field
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() {
			return app.thumb(field.String())
		}
	}
	return ""
}

func (app *App) getItemImageLarge(item interface{}) string {
	//TODO: Authorize field
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() {
			return app.largeImage(field.String())
		}
	}
	return ""
}

type namedIFace interface {
	GetName() string
}

func getItemName(item interface{}) string {
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
	return fmt.Sprintf("#%d", getItemID(item))
}

func (resourceData *resourceData) getItemDescription(item any, user *user, relatedResource *resourceData) string {
	var items []string
	itemsVal := reflect.ValueOf(item).Elem()

	//TODO: Authorize description field
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

func (app App) relationStringer(field Field, value reflect.Value, user *user) string {

	switch value.Kind() {
	case reflect.String:
		if field.tags["prago-type"] == "image" || field.tags["prago-type"] == "file" {
			return fmt.Sprintf("%dx", len(strings.Split(value.String(), ",")))
		}
		return value.String()
	case reflect.Int, reflect.Int32, reflect.Int64:
		if field.tags["prago-type"] == "relation" {
			if value.Int() <= 0 {
				return ""
			}
			field.relatedResource.resourceItemName(int64(value.Int()))
		}
		return fmt.Sprintf("%v", value.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	case reflect.Bool:
		if value.Bool() {
			return messages.Get(user.Locale, "yes_plain")
		}
		return messages.Get(user.Locale, "no_plain")
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Now()) {
			tm := value.Interface().(time.Time)
			showTime := false
			if field.tags["prago-type"] == "timestamp" {
				showTime = true
			}
			return messages.Timestamp(user.Locale, tm, showTime)
		}
	}
	return ""
}

func (resourceData *resourceData) resourceItemName(id int64) string {
	item := resourceData.query().ID(id)
	if item == nil {
		return fmt.Sprintf("%d - not found", id)
	}
	return getItemName(item)
}

func getItemID(item interface{}) int64 {
	if item == nil {
		return -1
	}

	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("ID")
	if field.IsValid() {
		return field.Int()
	}
	return -1
}
