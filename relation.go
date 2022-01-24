package prago

import (
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
	for _, v := range app.resources {
		v.createRelations()
	}
}

func (resource *Resource[T]) createRelations() {
	for _, field := range resource.fields {
		if field.tags["prago-type"] == "relation" {
			relatedResourceID := field.columnName
			if field.tags["prago-relation"] != "" {
				relatedResourceID = field.tags["prago-relation"]
			}
			field.relatedResource = resource.app.getResourceByID(relatedResourceID)

			//TODO: name can be set directly
			if field.tags["prago-name"] == "" {
				field.humanName = field.relatedResource.getNameFunction()
			}
			field.relatedResource.addRelation((*relatedField)(field))
		}
	}
}

func (field *relatedField) addURL(id int64) string {
	values := url.Values{}
	values.Add(field.columnName, fmt.Sprintf("%d", id))
	return field.resource.getURL("new?" + values.Encode())
}

func (field *relatedField) listURL(id int64) string {
	values := url.Values{}
	values.Add(field.columnName, fmt.Sprintf("%d", id))
	return field.resource.getURL("") + "?" + values.Encode()
}

func (field *relatedField) listName(locale string) string {
	ret := field.resource.getName(locale)
	fieldName := field.humanName(locale)
	referenceName := field.relatedResource.getName(locale)
	if fieldName != referenceName {
		ret += " – " + fieldName
	}
	return ret
}

func getRelationViewData(user *user, f *Field, value interface{}) interface{} {
	ret, _ := f.resource.getPreviewData(user, f, value.(int64))
	return ret
}

func (resource *Resource[T]) getPreviewData(user *user, f *Field, value int64) (*preview, error) {
	app := f.resource.getApp()
	if f.relatedResource == nil {
		return nil, fmt.Errorf("resource not found: %s", f.humanName("en"))
	}

	if !app.authorize(user, f.relatedResource.getPermissionView()) {
		return nil, fmt.Errorf("user is not authorized to view this item")
	}

	return f.relatedResource.getItemPreview(value, user, f.resource), nil
}

func (resource *Resource[T]) getItemPreview(id int64, user *user, relatedResource resourceIface) *preview {
	item := resource.Is("id", id).First()
	if item == nil {
		return nil
	}
	return resource.getPreview(item, user, nil)

}

func (resource *Resource[T]) getPreview(item *T, user *user, relatedResource resourceIface) *preview {
	var ret preview
	ret.ID = getItemID(item)
	ret.Name = getItemName(item)
	ret.URL = resource.getItemURL(item, "")
	ret.Image = resource.app.getItemImage(item)
	ret.Description = resource.getItemDescription(item, user, relatedResource)
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

func getItemName(item interface{}) string {
	//TODO: Authorize field
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
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

func (resource *Resource[T]) getItemDescription(item *T, user *user, relatedResource resourceIface) string {
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

	for _, v := range resource.fields {
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
		stringed := resource.app.relationStringer(*v, field, user)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.humanName(user.Locale), stringed))
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

func (resource *Resource[T]) resourceItemName(id int64) string {
	item := resource.Is("id", id).First()
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
