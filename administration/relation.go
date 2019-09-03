package administration

import (
	"fmt"
	"net/url"
	"reflect"

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
	var relationName string
	if f.Tags["prago-relation"] != "" {
		relationName = f.Tags["prago-relation"]
	} else {
		relationName = f.Name
	}

	r2 := resource.Admin.getResourceByName(relationName)
	if r2 == nil {
		fmt.Printf("Resource '%s' not found\n", relationName)
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

	return r2.itemToRelationData(item)
}

func (resource *Resource) itemToRelationData(item interface{}) *viewRelationData {
	var ret viewRelationData
	ret.ID = getItemID(item)
	ret.Name = getItemName(item)
	ret.URL = resource.GetItemURL(item, "")

	ret.Image = resource.Admin.getItemImage(item)
	ret.Description = getItemDescription(item)
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

func getItemDescription(item interface{}) string {
	if item != nil {
		itemsVal := reflect.ValueOf(item).Elem()
		field := itemsVal.FieldByName("Description")
		if field.IsValid() {
			ret := field.String()
			return utils.CropMarkdown(ret, 200)
		}
	}
	return ""
}

func getItemID(item interface{}) int64 {
	if item == nil {
		return 0
	}

	itemsVal := reflect.ValueOf(item).Elem()
	field := itemsVal.FieldByName("ID")
	if field.IsValid() {
		return field.Int()
	}
	return 0
}
