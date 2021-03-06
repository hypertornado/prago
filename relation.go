package prago

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type relation struct {
	resource *Resource
	field    string
	listName func(string) string
	listURL  func(id int64) string
	addURL   func(id int64) string
}

type viewRelationData struct {
	ID          int64
	Image       string
	URL         string
	Name        string
	Description string
}

func (app *App) initAllAutoRelations() {
	for _, v := range app.resources {
		v.initAutoRelations()
	}
}

func (resource *Resource) initAutoRelations() {
	for _, v := range resource.fieldArrays {
		if v.Tags["prago-type"] == "relation" {
			referenceName := v.Name
			relationFieldName := v.Tags["prago-relation"]
			if relationFieldName != "" {
				referenceName = relationFieldName
			}
			relatedResource := resource.app.getResourceByName(referenceName)
			if relatedResource == nil {
				panic("can't find reference resource: " + referenceName)
			}
			v.relatedResource = relatedResource

			if v.Tags["prago-name"] == "" {
				v.HumanName = (*relatedResource).name
			}

			relatedResource.relations = append(relatedResource.relations, relation{
				resource: resource,
				field:    v.Name,
				listName: createRelationNamingFunction(*v, *resource, *relatedResource),
				listURL:  createRelationListURL(*resource, *v),
				addURL:   createRelationAddURL(*resource, *v),
			})
		}
	}
}

func createRelationAddURL(resource Resource, field field) func(int64) string {
	return func(id int64) string {
		values := url.Values{}
		values.Add(field.ColumnName, fmt.Sprintf("%d", id))
		return resource.getURL("new?" + values.Encode())
	}
}

func createRelationListURL(resource Resource, field field) func(int64) string {
	return func(id int64) string {
		values := url.Values{}
		values.Add(field.ColumnName, fmt.Sprintf("%d", id))
		return resource.getURL("") + "?" + values.Encode()
	}
}

func createRelationNamingFunction(field field, resource Resource, referenceResource Resource) func(string) string {
	return func(lang string) string {
		ret := resource.name(lang)
		fieldName := field.HumanName(lang)
		referenceName := referenceResource.name(lang)
		if fieldName != referenceName {
			ret += " – " + fieldName
		}
		return ret
	}
}

func getRelationViewData(user *user, f field, value interface{}) interface{} {
	ret, _ := getRelationData(user, f, value)
	return ret
}

func getRelationData(user *user, f field, value interface{}) (*viewRelationData, error) {
	app := f.resource.app
	if f.relatedResource == nil {
		return nil, fmt.Errorf("resource not found: %s", f.Name)
	}

	if !app.authorize(user, f.relatedResource.canView) {
		return nil, fmt.Errorf("user is not authorized to view this item")
	}

	var item interface{}
	f.relatedResource.newItem(&item)

	intVal := value.(int64)
	if intVal <= 0 {
		return nil, fmt.Errorf("wrong value")
	}
	err := app.Query().WhereIs("id", intVal).Get(item)
	if err != nil {
		return nil, fmt.Errorf("can't find this item")
	}

	return f.relatedResource.itemToRelationData(item, user, nil), nil
}

func (resource *Resource) itemToRelationData(item interface{}, user *user, relatedResource *Resource) *viewRelationData {
	var ret viewRelationData
	ret.ID = getItemID(item)
	ret.Name = getItemName(item)
	ret.URL = resource.getItemURL(item, "")

	ret.Image = resource.app.getItemImage(item)
	ret.Description = resource.getItemDescription(item, user, relatedResource)
	return &ret
}

func (app *App) getItemImage(item interface{}) string {
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

func (resource *Resource) getItemDescription(item interface{}, user *user, relatedResource *Resource) string {
	var items []string

	itemsVal := reflect.ValueOf(item).Elem()

	//TODO: Authorize descriptio field
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

	for _, v := range resource.fieldArrays {
		if v.Name == "ID" || v.Name == "Name" || v.Name == "Description" {
			continue
		}
		if !v.authorizeView(user) {
			continue
		}

		rr := v.relatedResource
		if rr != nil && relatedResource != nil && rr.id == relatedResource.id {
			continue
		}

		field := itemsVal.FieldByName(v.Name)
		stringed := resource.app.relationStringer(*v, field, user)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.HumanName(user.Locale), stringed))
		}
	}
	ret := strings.Join(items, " · ")
	return cropMarkdown(ret, 500)
}

func (app App) relationStringer(field field, value reflect.Value, user *user) string {
	switch value.Kind() {
	case reflect.String:
		if field.Tags["prago-type"] == "image" || field.Tags["prago-type"] == "file" {
			return fmt.Sprintf("%dx", len(strings.Split(value.String(), ",")))
		}
		return value.String()
	case reflect.Int, reflect.Int32, reflect.Int64:
		if field.Tags["prago-type"] == "relation" {
			if value.Int() <= 0 {
				return ""
			}
			rr := field.relatedResource

			var item interface{}
			rr.newItem(&item)
			err := rr.app.Query().WhereIs("id", int64(value.Int())).Get(item)
			if err != nil {
				return fmt.Sprintf("%d", value.Int())
			}
			return getItemName(item)
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
			if field.Tags["prago-type"] == "timestamp" {
				showTime = true
			}
			return messages.Timestamp(user.Locale, tm, showTime)
		}
	}
	return ""
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
