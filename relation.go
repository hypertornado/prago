package prago

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type relatedField Field

func (app *App) initRelations() {
	for _, resource := range app.resources {
		resource.createRelations()
	}
}

func (resource *Resource) createRelations() {
	for _, field := range resource.fields {
		if field.tags["prago-type"] == "relation" || field.tags["prago-type"] == "multirelation" {

			relatedResourceID := field.id
			if field.tags["prago-relation"] != "" {
				relatedResourceID = field.tags["prago-relation"]
			}
			field.relatedResource = resource.app.getResourceByID(relatedResourceID)
			if field.relatedResource == nil {
				panic(fmt.Sprintf("can't find related resource by id '%s', resource: %s, field: %s", relatedResourceID, resource.id, field.id))
			}

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
	return field.resource.getURL("list") + "?" + values.Encode()
}

func (field *relatedField) listName(locale string) string {
	f := (*Field)(field)
	ret := f.GetManuallySetPluralName(locale)
	if ret != "" {
		return fmt.Sprintf("%s – %s", field.resource.pluralName(locale), ret)
	}
	return field.resource.pluralName(locale)
}

func (app App) relationStringer(field Field, value reflect.Value, userData UserData) string {

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

			id := int64(value.Int())
			item := field.relatedResource.query(context.TODO()).ID(id)
			if item == nil {
				return fmt.Sprintf("%d - not found", id)
			}
			return field.relatedResource.previewer(userData, item).Name()
		}
		return fmt.Sprintf("%v", value.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	case reflect.Bool:
		if value.Bool() {
			return messages.Get(userData.Locale(), "yes_plain")
		}
		return messages.Get(userData.Locale(), "no_plain")
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Now()) {
			tm := value.Interface().(time.Time)
			showTime := false
			if field.tags["prago-type"] == "timestamp" {
				showTime = true
			}
			return messages.Timestamp(userData.Locale(), tm, showTime)
		}
	}
	return ""
}
