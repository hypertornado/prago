package prago

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"sync"
	"time"
)

//https://github.com/golang/go/issues/49085

var resourceMapMutex = &sync.RWMutex{}

type Resource struct {
	app *App
	id  string

	icon string

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

	actions     []*Action
	itemActions []*Action

	relations []*relatedField

	resourceController *controller

	orderByColumn       string
	orderDesc           bool
	defaultItemsPerPage int64

	typ reflect.Type

	fields     []*Field
	fieldMap   map[string]*Field
	orderField *Field

	resourceBoard *Board
	parentBoard   *Board

	previewFn func(any) string

	itemStats []*itemStat
}

func NewResource[T any](app *App) *Resource {
	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()
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

	ret := &Resource{
		app:  app,
		id:   columnName(defaultName),
		icon: iconResource,

		singularName: unlocalized(defaultName),
		pluralName:   unlocalized(defaultName),

		activityLog: true,

		canView:   sysadminPermission,
		canCreate: loggedPermission,
		canUpdate: loggedPermission,
		canDelete: loggedPermission,
		canExport: loggedPermission,

		resourceController: app.adminController.subController(),

		defaultItemsPerPage: 100,

		typ: typ,

		parentBoard: app.MainBoard,

		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := ret.newField(typ.Field(i), i)
			if field.tags["prago-type"] == "order" {
				ret.orderField = field
			}
			ret.fields = append(ret.fields, field)
			ret.fieldMap[field.id] = field
		}
	}

	app.resources = append(app.resources, ret)
	app.resourceMap[ret.typ] = ret
	app.resourceNameMap[ret.id] = ret

	initResource(ret)

	ret.resourceBoard = &Board{
		app:            app,
		parentResource: ret,
	}

	ret.resourceBoard.MainDashboard = &Dashboard{
		name:  unlocalized(""),
		board: ret.resourceBoard,
	}

	statsDashboard := ret.resourceBoard.Dashboard(unlocalized("Statistiky"))
	statsDashboard.Figure(unlocalized(""), ret.canView).Value(func(r *Request) int64 {
		c, _ := ret.query(context.Background()).count()
		return c
	}).Unit(unlocalized("položek"))

	ret.orderByColumn, ret.orderDesc = ret.getDefaultOrder()
	return ret
}

func GetResource[T any](app *App) *Resource {
	resourceMapMutex.RLock()
	defer resourceMapMutex.RUnlock()

	var item T
	itemTyp := reflect.TypeOf(item)
	ret, ok := app.resourceMap[itemTyp]
	if !ok {
		return nil
	}
	return ret
}

func (resource *Resource) isItPointerToResourceItem(item any) bool {
	if item == nil {
		return false
	}
	return reflect.PointerTo(resource.typ) == reflect.TypeOf(item)
}

func (resource *Resource) addRelation(field *relatedField) {
	resource.relations = append(resource.relations, field)
}

func (resource *Resource) getID() string {
	return resource.id
}

func (resource *Resource) getResourceControl() *controller {
	return resource.resourceController
}

func CreateItem[T any](app *App, item *T) error {
	return CreateItemWithContext(context.Background(), app, item)
}

func CreateItemWithContext[T any](ctx context.Context, app *App, item *T) error {
	resource := GetResource[T](app)
	return resource.create(ctx, item)
}

func (resource *Resource) create(ctx context.Context, item any) error {
	resource.setTimestamp(item, "CreatedAt")
	resource.setTimestamp(item, "UpdatedAt")
	return resource.createItem(ctx, item, false)
}

func UpdateItem[T any](app *App, item *T) error {
	return UpdateItemWithContext[T](context.Background(), app, item)
}

func UpdateItemWithContext[T any](ctx context.Context, app *App, item *T) error {
	resource := GetResource[T](app)
	return resource.update(ctx, item)
}

func (resource *Resource) update(ctx context.Context, item any) error {
	resource.setTimestamp(item, "UpdatedAt")
	return resource.saveItem(ctx, item, false)
}

func Replace[T any](ctx context.Context, app *App, item *T) error {
	resource := GetResource[T](app)
	resource.setTimestamp(item, "CreatedAt")
	resource.setTimestamp(item, "UpdatedAt")
	return resource.replaceItem(ctx, item, false)
}

func (resource *Resource) setTimestamp(item any, fieldName string) {
	val := reflect.ValueOf(item).Elem()
	fieldVal := val.FieldByName(fieldName)
	timeVal := reflect.ValueOf(time.Now())
	if fieldVal.IsValid() &&
		fieldVal.CanSet() &&
		fieldVal.Type() == timeVal.Type() {
		fieldVal.Set(timeVal)
	}
}

func DeleteItem[T any](app *App, id int64) error {
	return DeleteItemWithContext[T](context.Background(), app, id)
}

func DeleteItemWithContext[T any](ctx context.Context, app *App, id int64) error {
	resource := GetResource[T](app)
	return resource.delete(ctx, id)
}

func (resource *Resource) delete(ctx context.Context, id int64) error {
	q := resource.query(ctx).Is("id", id)
	count, err := q.delete()
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

func (resource *Resource) Name(singularName, pluralName func(string) string) *Resource {
	resource.singularName = singularName
	resource.pluralName = pluralName
	return resource
}

func PreviewURLFunction[T any](app *App, fn func(*T) string) {
	resource := GetResource[T](app)
	resource.previewURLFunction(func(a any) string {
		return fn(a.(*T))
	})
}

func (resource *Resource) previewURLFunction(fn func(any) string) {
	resource.previewFn = fn
}

func (resource *Resource) Icon(icon string) *Resource {
	resource.icon = icon
	return resource

}

func (resource *Resource) ItemsPerPage(itemsPerPage int64) *Resource {
	resource.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *Resource) PermissionView(permission Permission) *Resource {
	must(resource.app.validatePermission(permission))
	resource.canView = permission
	return resource
}

func (resource *Resource) PermissionUpdate(permission Permission) *Resource {
	must(resource.app.validatePermission(permission))
	if resource.canCreate == loggedPermission {
		resource.canCreate = permission
	}
	if resource.canDelete == loggedPermission {
		resource.canDelete = permission
	}
	resource.canUpdate = permission
	return resource
}

func (resource *Resource) PermissionCreate(permission Permission) *Resource {
	must(resource.app.validatePermission(permission))
	resource.canCreate = permission
	return resource
}

func (resource *Resource) PermissionDelete(permission Permission) *Resource {
	must(resource.app.validatePermission(permission))
	resource.canDelete = permission
	return resource
}

func (resource *Resource) PermissionExport(permission Permission) *Resource {
	must(resource.app.validatePermission(permission))
	resource.canExport = permission
	return resource
}

func (resource *Resource) Validation(validation Validation) *Resource {
	resource.addValidation(validation)
	return resource
}

func (resource *Resource) Dashboard(name func(string) string) *Dashboard {
	return resource.resourceBoard.Dashboard(name)
}

func (resource *Resource) addValidation(validation Validation) {
	resource.validations = append(resource.validations, validation)
}

func (resource *Resource) DeleteValidation(validation Validation) *Resource {

	resource.deleteValidations = append(resource.deleteValidations, validation)
	return resource
}

func (resource *Resource) Board(board *Board) *Resource {
	resource.parentBoard = board
	return resource
}

func (resource *Resource) getItemURL(item interface{}, suffix string, userData UserData) string {
	ret := resource.getURL(fmt.Sprintf("%d", resource.previewer(userData, item).ID()))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func (app *App) getResourceByID(name string) *Resource {
	resourceMapMutex.RLock()
	defer resourceMapMutex.RUnlock()
	return app.resourceNameMap[columnName(name)]
}

func initResource(resource *Resource) {
	resource.resourceController.addAroundAction(func(request *Request, next func()) {
		if !request.Authorize(resource.canView) {
			renderErrorPage(request, 403)
		} else {
			next()
		}
	})
}

func (resource *Resource) getURL(suffix string) string {
	url := resource.id
	if len(suffix) > 0 {
		url += "/" + suffix
	}
	return resource.app.getAdminURL(url)
}

func (resource *Resource) cachedCountName() string {
	return fmt.Sprintf("prago-resource_count-%s", resource.id)
}

func (resource *Resource) getCachedCount() int64 {
	return loadCache(resource.app.cache, resource.cachedCountName(), func() int64 {
		count, _ := resource.countAllItems(context.Background())
		return count
	})
}

func (resource *Resource) updateCachedCount() error {
	resource.app.cache.forceLoad(resource.cachedCountName(), func() interface{} {
		count, _ := resource.countAllItems(context.Background())
		return count
	})
	return nil
}
