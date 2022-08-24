package prago

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"time"

	"github.com/hypertornado/prago/pragelastic"
)

//https://github.com/golang/go/issues/49085

type Resource[T any] struct {
	data               *resourceData
	previewURLFunction func(*T) string
}

type resourceData struct {
	app *App
	id  string

	singularName func(locale string) string
	pluralName   func(locale string) string

	activityLog bool

	canView   Permission
	canCreate Permission
	canUpdate Permission
	canDelete Permission
	canExport Permission

	validations       []Validation
	deleteValidations []Validation

	actions     []ActionIface
	itemActions []ActionIface

	relations []*relatedField

	resourceController *controller

	orderByColumn       string
	orderDesc           bool
	defaultItemsPerPage int64

	typ reflect.Type

	quickActions []quickActionIface

	fields     []*Field
	fieldMap   map[string]*Field
	orderField *Field
}

func NewResource[T any](app *App) *Resource[T] {
	var item T
	typ := reflect.TypeOf(item)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("item is not a structure, but " + typ.Kind().String()))
	}

	_, typFound := app.resourceMap[typ]
	if typFound {
		panic(fmt.Errorf("resource with type %s already created", typ))
	}

	defaultName := typ.Name()

	data := &resourceData{
		app: app,
		id:  columnName(defaultName),

		singularName: unlocalized(defaultName),
		pluralName:   unlocalized(defaultName),

		activityLog: true,

		canView:   sysadminPermission,
		canCreate: loggedPermission,
		canUpdate: loggedPermission,
		canDelete: loggedPermission,
		canExport: loggedPermission,

		resourceController: app.adminController.subController(),

		defaultItemsPerPage: 200,

		typ: typ,

		fieldMap: make(map[string]*Field),
	}

	ret := &Resource[T]{
		data: data,
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := ret.newField(typ.Field(i), i)
			if field.tags["prago-type"] == "order" {
				ret.data.orderField = field
			}
			ret.data.fields = append(ret.data.fields, field)
			ret.data.fieldMap[field.id] = field
		}
	}

	app.resources = append(app.resources, ret)
	app.resourceMap[ret.data.typ] = ret
	app.resourceNameMap[ret.data.id] = ret

	initResource(ret.data)

	ret.data.orderByColumn, ret.data.orderDesc = ret.data.getDefaultOrder()
	return ret
}

func GetResource[T any](app *App) *Resource[T] {
	var item T
	itemTyp := reflect.TypeOf(item)
	ret, ok := app.resourceMap[itemTyp]
	if !ok {
		return nil
	}
	return ret.(*Resource[T])

}

type resourceIface interface {
	initDefaultResourceActions()
	initDefaultResourceAPIs()
	//createRelations()
	//addValidation(validation Validation)

	//getCachedCount() int64
	//addRelation(*relatedField)

	//getResourceControl() *controller
	//getID() string

	//bindActions()

	getData() *resourceData

	//getPermissionView() Permission
	//getPermissionCreate() Permission
	//getPermissionUpdate() Permission
	//getPermissionDelete() Permission
	//getPermissionExport() Permission

	//getSingularNameFunction() func(string) string
	//getPluralNameFunction() func(string) string

	//getApp() *App

	//unsafeDropTable() error
	//migrate(bool) error

	//getURL(suffix string) string

	//getPreviewData(user *user, f *Field, value int64) (*preview, error)
	//getItemPreview(id int64, user *user, relatedResource resourceIface) *preview
	//getPreviews(listRequest relationListRequest, user *user) []*preview

	//getnavigation(action *Action, request *Request) navigation

	importSearchData(*pragelastic.BulkUpdater[searchItem]) error
	//resourceItemName(id int64) string
	//itemWithRelationCount(fieldName string, id int64) int64

	//runQuickAction(actionName string, itemID int64, request *Request) error
}

func (resource *Resource[T]) getData() *resourceData {
	return resource.data
}

/*func (resource *Resource[T]) getApp() *App {
	return resource.data.app
}*/

/*func (resource *Resource[T]) getSingularNameFunction() func(string) string {
	return resource.data.singularName
}*/

/*func (resource *Resource[T]) getPluralNameFunction() func(string) string {
	return resource.data.pluralName
}*/

/*func (resource *Resource[T]) getPermissionView() Permission {
	return resource.data.canView
}*/

/*func (resource *Resource[T]) getPermissionCreate() Permission {
	return resource.data.canCreate
}*/

/*func (resource *Resource[T]) getPermissionUpdate() Permission {
	return resource.data.canUpdate
}*/

/*func (resource *Resource[T]) getPermissionDelete() Permission {
	return resource.data.canDelete
}*/

/*func (resource *Resource[T]) getPermissionExport() Permission {
	return resource.data.canDelete
}*/

func (resourceData *resourceData) addRelation(field *relatedField) {
	resourceData.relations = append(resourceData.relations, field)
}

func (resourceData *resourceData) getID() string {
	return resourceData.id
}

func (resourceData *resourceData) getResourceControl() *controller {
	return resourceData.resourceController
}

/*func (resource *Resource[T]) isOrderDesc() bool {
	return resource.data.orderDesc
}*/

/*func (resource *Resource[T]) getOrderByColumn() string {
	return resource.data.orderByColumn
}*/

func (resource *Resource[T]) Is(name string, value interface{}) *Query[T] {
	return resource.Query().Is(name, value)
}

func (resource *Resource[T]) Create(item *T) error {
	resource.setTimestamp(item, "CreatedAt")
	resource.setTimestamp(item, "UpdatedAt")
	return resource.data.createItem(item, false)
}

func (resource *Resource[T]) Update(item *T) error {
	resource.setTimestamp(item, "UpdatedAt")
	return resource.data.saveItem(item, false)
}

func (resource *Resource[T]) Replace(item *T) error {
	resource.setTimestamp(item, "CreatedAt")
	resource.setTimestamp(item, "UpdatedAt")
	return resource.data.replaceItem(item, false)
}

func (resource *Resource[T]) setTimestamp(item *T, fieldName string) {
	val := reflect.ValueOf(item).Elem()
	fieldVal := val.FieldByName(fieldName)
	timeVal := reflect.ValueOf(time.Now())
	if fieldVal.IsValid() &&
		fieldVal.CanSet() &&
		fieldVal.Type() == timeVal.Type() {
		fieldVal.Set(timeVal)
	}
}

func (resource *Resource[T]) Delete(id int64) error {
	q := resource.Is("id", id)
	count, err := q.listQuery.delete()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("no item deleted")
	}
	if count > 1 {
		return fmt.Errorf("more then one item deleted: %d items deleted", count)
	}
	return nil
}

func (resource *Resource[T]) Name(singularName, pluralName func(string) string) *Resource[T] {
	resource.data.singularName = singularName
	resource.data.pluralName = pluralName
	return resource
}

func (resource *Resource[T]) PreviewURLFunction(fn func(*T) string) *Resource[T] {
	resource.previewURLFunction = fn
	return resource
}

func (resource *Resource[T]) ItemsPerPage(itemsPerPage int64) *Resource[T] {
	resource.data.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *Resource[T]) PermissionView(permission Permission) *Resource[T] {
	must(resource.data.app.validatePermission(permission))
	resource.data.canView = permission
	return resource
}

func (resource *Resource[T]) PermissionUpdate(permission Permission) *Resource[T] {
	must(resource.data.app.validatePermission(permission))
	if resource.data.canCreate == loggedPermission {
		resource.data.canCreate = permission
	}
	if resource.data.canDelete == loggedPermission {
		resource.data.canDelete = permission
	}
	resource.data.canUpdate = permission
	return resource
}

func (resource *Resource[T]) PermissionCreate(permission Permission) *Resource[T] {
	must(resource.data.app.validatePermission(permission))
	resource.data.canCreate = permission
	return resource
}

func (resource *Resource[T]) PermissionDelete(permission Permission) *Resource[T] {
	must(resource.data.app.validatePermission(permission))
	resource.data.canDelete = permission
	return resource
}

func (resource *Resource[T]) PermissionExport(permission Permission) *Resource[T] {
	must(resource.data.app.validatePermission(permission))
	resource.data.canExport = permission
	return resource
}

func (resource *Resource[T]) Validation(validation Validation) *Resource[T] {
	resource.data.addValidation(validation)
	return resource
}

func (resourceData *resourceData) addValidation(validation Validation) {
	resourceData.validations = append(resourceData.validations, validation)
}

func (resource *Resource[T]) DeleteValidation(validation Validation) *Resource[T] {
	resource.data.deleteValidations = append(resource.data.deleteValidations, validation)
	return resource
}

func (resourceData *resourceData) allowsMultipleActions(user *user) (ret bool) {
	if resourceData.app.authorize(user, resourceData.canDelete) {
		ret = true
	}
	if resourceData.app.authorize(user, resourceData.canUpdate) {
		ret = true
	}
	return ret
}

func (resourceData *resourceData) getMultipleActions(user *user) (ret []listMultipleAction) {
	if !resourceData.allowsMultipleActions(user) {
		return nil
	}

	if resourceData.app.authorize(user, resourceData.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if resourceData.app.authorize(user, resourceData.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	if resourceData.app.authorize(user, resourceData.canDelete) {
		ret = append(ret, listMultipleAction{
			ID:       "delete",
			Name:     "Smazat",
			IsDelete: true,
		})
	}
	ret = append(ret, listMultipleAction{
		ID:   "cancel",
		Name: "Storno",
	})
	return
}

func (resourceData *resourceData) getItemURL(item interface{}, suffix string) string {
	ret := resourceData.getURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func (app *App) getResourceByID(name string) resourceIface {
	return app.resourceNameMap[columnName(name)]
}

func initResource(resourceData *resourceData) {
	resourceData.resourceController.addAroundAction(func(request *Request, next func()) {
		if !resourceData.app.authorize(request.user, resourceData.canView) {
			render403(request)
		} else {
			next()
		}
	})
}

func (resourceData *resourceData) getURL(suffix string) string {
	url := resourceData.id
	if len(suffix) > 0 {
		url += "/" + suffix
	}
	return resourceData.app.getAdminURL(url)
}

func (resourceData *resourceData) cachedCountName() string {
	return fmt.Sprintf("prago-resource_count-%s", resourceData.id)
}

func (resourceData *resourceData) getCachedCount() int64 {
	return loadCache(resourceData.app.cache, resourceData.cachedCountName(), func() int64 {
		count, _ := resourceData.countAllItems(false)
		return count
	})
}

func (resourceData *resourceData) updateCachedCount() error {
	resourceData.app.cache.forceLoad(resourceData.cachedCountName(), func() interface{} {
		count, _ := resourceData.countAllItems(false)
		return count
	})
	return nil
}
