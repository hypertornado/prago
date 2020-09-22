package administration

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
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

func (resource *Resource) initAutoRelations() {
	for _, v := range resource.fieldArrays {
		if v.Tags["prago-type"] == "relation" {
			referenceName := v.Name
			relationFieldName := v.Tags["prago-relation"]
			if relationFieldName != "" {
				referenceName = relationFieldName
			}
			referenceResource := resource.Admin.getResourceByName(referenceName)

			if v.Tags["prago-description"] == "" {
				v.HumanName = (*referenceResource).HumanName
			}

			referenceResource.autoRelations = append(referenceResource.autoRelations, relation{
				resource: resource,
				field:    v.Name,
				listName: createRelationNamingFunction(*v, *resource, *referenceResource),
				listURL:  createRelationListURL(*resource, *v),
				addURL:   createRelationAddURL(*resource, *v),
			})
		}
	}
}

func createRelationAddURL(resource Resource, field Field) func(int64) string {
	return func(id int64) string {
		values := url.Values{}
		values.Add(field.ColumnName, fmt.Sprintf("%d", id))
		return resource.GetURL("new?" + values.Encode())
	}
}

func createRelationListURL(resource Resource, field Field) func(int64) string {
	return func(id int64) string {
		values := url.Values{}
		values.Add(field.ColumnName, fmt.Sprintf("%d", id))
		return resource.GetURL("") + "?" + values.Encode()
	}
}

func createRelationNamingFunction(field Field, resource Resource, referenceResource Resource) func(string) string {
	return func(lang string) string {
		ret := resource.HumanName(lang)
		fieldName := field.HumanName(lang)
		referenceName := referenceResource.HumanName(lang)
		if fieldName != referenceName {
			ret += " – " + fieldName
		}
		return ret
	}
}

func getRelationViewData(resource Resource, user User, f Field, value interface{}) interface{} {
	return getRelationData(resource, user, f, value)
}

func getRelationData(resource Resource, user User, f Field, value interface{}) *viewRelationData {
	r2 := f.getRelatedResource(*resource.Admin)
	if r2 == nil {
		fmt.Printf("Resource not found: %s\n", f.Name)
		return nil
	}

	if !resource.Admin.Authorize(user, r2.CanView) {
		fmt.Printf("User is not authorized to view this item\n")
		return nil
	}

	var item interface{}
	r2.newItem(&item)

	intVal := value.(int64)
	if intVal <= 0 {
		return nil
	}
	err := resource.Admin.Query().WhereIs("id", intVal).Get(item)
	if err != nil {
		fmt.Printf("Can't find this item\n")
		return nil
	}

	return r2.itemToRelationData(item, user, nil)
}

func (resource *Resource) itemToRelationData(item interface{}, user User, relatedResource *Resource) *viewRelationData {
	var ret viewRelationData
	ret.ID = getItemID(item)
	ret.Name = getItemName(item)
	ret.URL = resource.GetItemURL(item, "")

	ret.Image = resource.Admin.getItemImage(item)
	ret.Description = resource.getItemDescription(item, user, relatedResource)
	return &ret
}

func (admin *Administration) getItemImage(item interface{}) string {
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		field := itemsVal.FieldByName("Image")
		if field.IsValid() {
			return admin.thumb(field.String())
		}
	}
	return ""
}

func (admin *Administration) itemHasImage(item interface{}) bool {
	if item == nil {
		return false
	}
	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("Image")
	if field.IsValid() {
		return true
	}
	return false
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

func (resource *Resource) getItemDescription(item interface{}, user User, relatedResource *Resource) string {
	var items []string

	itemsVal := reflect.ValueOf(item).Elem()

	if item != nil {
		field := itemsVal.FieldByName("Description")
		if field.IsValid() {
			ret := field.String()
			croped := utils.CropMarkdown(ret, 200)
			if croped != "" {
				items = append(items, croped)
			}
		}
	}

	//TODO: can show field? add access system for fields
	for _, v := range resource.fieldArrays {
		if v.Name == "ID" || v.Name == "Name" || v.Name == "Description" || v.Name == "OrderPosition" {
			continue
		}

		rr := v.getRelatedResource(*resource.Admin)
		if rr != nil && relatedResource != nil && rr.TableName == relatedResource.TableName {
			continue
		}

		field := itemsVal.FieldByName(v.Name)
		stringed := resource.Admin.relationStringer(*v, field, user)
		if stringed != "" {
			items = append(items, fmt.Sprintf("%s: %s", v.HumanName(user.Locale), stringed))
		}
	}
	ret := strings.Join(items, " · ")
	return utils.CropMarkdown(ret, 500)
}

func (admin Administration) relationStringer(field Field, value reflect.Value, user User) string {
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
			rr := field.getRelatedResource(admin)

			var item interface{}
			rr.newItem(&item)
			err := rr.Admin.Query().WhereIs("id", int64(value.Int())).Get(item)
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
			return messages.Messages.Get(user.Locale, "yes_plain")
		} else {
			return messages.Messages.Get(user.Locale, "no_plain")
		}
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Now()) {
			tm := value.Interface().(time.Time)
			showTime := false
			if field.Tags["prago-type"] == "timestamp" {
				showTime = true
			}
			return messages.Messages.Timestamp(user.Locale, tm, showTime)
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
