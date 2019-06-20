package administration

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
)

type relation struct {
	resource *Resource
	field    string
	addName  func(string) string
}

type viewRelationData struct {
	ID          int64
	Image       string
	URL         string
	Name        string
	Description string
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
		fmt.Errorf("Resource '%s' not found", relationName)
		return nil
	}

	if !resource.Admin.Authorize(user, r2.CanView) {
		fmt.Errorf("User is not authorized to view this item")
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
		fmt.Errorf("Can't find this item")
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

func (resource *Resource) AddRelation(relatedResource *Resource, field string, addName func(string) string) {
	resource.relations = append(resource.relations, relation{relatedResource, field, addName})
}

func (resource *Resource) bindRelationActions(r relation) {
	action := Action{
		Name:       r.resource.HumanName,
		URL:        r.resource.ID,
		Permission: r.resource.CanView,
		Handler: func(resource Resource, request prago.Request, user User) {
			listData, err := r.resource.getListHeader(user)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			listData.PrefilterField = r.field
			listData.PrefilterValue = request.Params().Get("id")

			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))

			navigation := resource.Admin.getItemNavigation(resource, user, item, r.resource.ID)
			navigation.Wide = true

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	}
	resource.AddItemAction(action)

	addAction := Action{
		Name:       r.addName,
		URL:        r.addURL(),
		Permission: r.resource.CanView,
		Handler: func(resource Resource, request prago.Request, user User) {
			relatedResource := *r.resource

			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)
			var item interface{}
			resource.newItem(&item)
			err = resource.Admin.Query().WhereIs("id", int64(id)).Get(item)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			var values = make(url.Values)
			values.Set(r.field, request.Params().Get("id"))

			var newItem interface{}
			relatedResource.newItem(&newItem)
			relatedResource.bindData(&newItem, user, values, defaultEditabilityFilter)
			form, err := relatedResource.getForm(newItem, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "../" + r.resource.ID
			form.Action = resource.Admin.Prefix + "/" + r.resource.ID
			form.AddSubmit("_submit", messages.Messages.Get(user.Locale, "admin_create"))
			AddCSRFToken(form, request)

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getItemNavigation(resource, user, item, r.addURL()),
				PageTemplate: "admin_form",
				PageData:     form,
			})
		},
	}

	resource.AddItemAction(addAction)
}

func (r relation) addURL() string {
	return "add-" + r.resource.ID
}
