package prago

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"sync"
)

//https://github.com/golang/go/issues/49085

var resourceMapMutex = &sync.RWMutex{}

type Resource struct {
	app *App
	id  string

	icon string

	hasImage bool

	singularName func(locale string) string
	pluralName   func(locale string) string

	activityLog bool

	canView   Permission
	canCreate Permission
	canUpdate Permission
	canDelete Permission
	canExport Permission

	updateValidations []func(any, Validation, UserData)
	deleteValidations []func(any, Validation, UserData)

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

	defaultValues map[string]func(*Request) string

	multipleActions []*MultipleItemAction

	customSearchFunctions []func(q string, userData UserData) []*Preview
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

	ret.resourceBoard.mainDashboard = &Dashboard{
		name:  unlocalized(""),
		board: ret.resourceBoard,
	}

	ret.orderByColumn, ret.orderDesc = ret.getDefaultOrder()
	return ret
}

func (resource *Resource) afterInit() {
	statsDashboard := resource.resourceBoard.Dashboard(unlocalized("Statistiky"))
	statsDashboard.Figure(unlocalized(""), resource.canView).Value(func(r *Request) int64 {
		c, _ := resource.query(context.Background()).count()
		return c
	}).Unit(unlocalized("položek")).URL(fmt.Sprintf("/admin/%s/list", resource.id))

	if resource.activityLog {
		statsDashboard.Figure(unlocalized(""), resource.canUpdate).Value(func(request *Request) int64 {
			q := resource.app.activityLogResource.query(context.Background())
			c, _ := q.Is("resourcename", resource.id).count()
			return c
		}).Unit(unlocalized("úprav")).URL(fmt.Sprintf("/admin/%s/history", resource.id))
	}

}

func getResource[T any](app *App) *Resource {
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

func (resource *Resource) DefaultValue(fieldName string, fn func(*Request) string) {
	fieldName = columnName(fieldName)
	if resource.fieldMap[fieldName] == nil {
		panic("can't find field name " + fieldName)
	}

	if resource.defaultValues == nil {
		resource.defaultValues = map[string]func(*Request) string{}
	}

	resource.defaultValues[fieldName] = fn
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

func PreviewURLFunction[T any](app *App, fn func(*T) string) {
	resource := getResource[T](app)
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

func (resource *Resource) Dashboard(name func(string) string) *Dashboard {
	return resource.resourceBoard.Dashboard(name)
}

func (resource *Resource) addUpdateValidation(validation func(any, Validation, UserData)) {
	resource.updateValidations = append(resource.updateValidations, validation)
}

func ValidateUpdate[T any](app *App, fn func(item *T, validation Validation, userData UserData)) {
	resource := getResource[T](app)
	resource.addUpdateValidation(func(item any, v Validation, userData UserData) {
		fn(item.(*T), v, userData)
	})
}

func ValidateDelete[T any](app *App, fn func(item *T, validation Validation, userData UserData)) {
	resource := getResource[T](app)
	resource.addDeleteValidation(func(a any, v Validation, userData UserData) {
		fn(a.(*T), v, userData)
	})

}

func (resource *Resource) addDeleteValidation(validation func(any, Validation, UserData)) *Resource {
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
		count := resource.countAllItems()
		return count
	})
}

func (resource *Resource) updateCachedCount() error {
	resource.app.cache.forceLoad(resource.cachedCountName(), func() interface{} {
		count := resource.countAllItems()
		return count
	})
	return nil
}

func (resource *Resource) forEach(ctx context.Context, handler func(any) error) error {
	var lastID int64
	var chunkSize = 100

	for {
		items, err := resource.query(ctx).Order("id").where("id > ?", lastID).Limit(int64(chunkSize)).list()
		if err != nil {
			return err
		}
		itemsCount := reflect.ValueOf(items).Len()
		if itemsCount == 0 {
			break
		}
		for i := range itemsCount {
			item := reflect.ValueOf(items).Index(i)
			val := item.Elem()
			id := val.FieldByName("ID").Int()
			lastID = id
			err := handler(item.Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ForEach[T any](app *App, ctx context.Context, handler func(*T) error) error {
	resource := getResource[T](app)
	return resource.forEach(ctx, func(a any) error {
		return handler(a.(*T))
	})
}
