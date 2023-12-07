package prago

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"time"
)

//https://github.com/golang/go/issues/49085

type Resource[T any] struct {
	data *resourceData
}

type resourceData struct {
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

	//quickActions []*quickActionData

	fields     []*Field
	fieldMap   map[string]*Field
	orderField *Field

	resourceBoard *Board
	parentBoard   *Board

	previewURLFunction func(any) string
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

		defaultItemsPerPage: 100,

		typ: typ,

		parentBoard: app.MainBoard,

		fieldMap: make(map[string]*Field),
	}

	ret := &Resource[T]{
		data: data,
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := ret.data.newField(typ.Field(i), i)
			if field.tags["prago-type"] == "order" {
				ret.data.orderField = field
			}
			ret.data.fields = append(ret.data.fields, field)
			ret.data.fieldMap[field.id] = field
		}
	}

	app.resources = append(app.resources, ret.data)
	app.resourceMap[ret.data.typ] = ret.data
	app.resourceNameMap[ret.data.id] = ret.data

	initResource(ret.data)

	ret.data.resourceBoard = &Board{
		app:            app,
		parentResource: ret.data,
	}

	ret.data.resourceBoard.MainDashboard = &Dashboard{
		name:  unlocalized(""),
		board: ret.data.resourceBoard,
	}

	statsDashboard := ret.data.resourceBoard.Dashboard(unlocalized("Statistiky"))
	statsDashboard.Figure(unlocalized(""), ret.data.canView).Value(func(r *Request) int64 {
		c, _ := ret.data.query(context.Background()).count()
		return c
	}).Unit(unlocalized("položek"))

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
	return &Resource[T]{
		data: ret,
	}

}

func (resourceData *resourceData) addRelation(field *relatedField) {
	resourceData.relations = append(resourceData.relations, field)
}

func (resourceData *resourceData) getID() string {
	return resourceData.id
}

func (resourceData *resourceData) getResourceControl() *controller {
	return resourceData.resourceController
}

func (resource *Resource[T]) Create(ctx context.Context, item *T) error {
	return resource.data.Create(ctx, item)
}

func (resourceData *resourceData) Create(ctx context.Context, item any) error {
	resourceData.setTimestamp(item, "CreatedAt")
	resourceData.setTimestamp(item, "UpdatedAt")
	return resourceData.createItem(ctx, item, false)
}

func (resource *Resource[T]) Update(ctx context.Context, item *T) error {
	return resource.data.Update(ctx, item)
}

func (resourceData *resourceData) Update(ctx context.Context, item any) error {
	resourceData.setTimestamp(item, "UpdatedAt")
	return resourceData.saveItem(ctx, item, false)
}

func (resource *Resource[T]) Replace(ctx context.Context, item *T) error {
	resource.data.setTimestamp(item, "CreatedAt")
	resource.data.setTimestamp(item, "UpdatedAt")
	return resource.data.replaceItem(ctx, item, false)
}

func (resourceData *resourceData) setTimestamp(item any, fieldName string) {
	val := reflect.ValueOf(item).Elem()
	fieldVal := val.FieldByName(fieldName)
	timeVal := reflect.ValueOf(time.Now())
	if fieldVal.IsValid() &&
		fieldVal.CanSet() &&
		fieldVal.Type() == timeVal.Type() {
		fieldVal.Set(timeVal)
	}
}

func (resource *Resource[T]) Delete(ctx context.Context, id int64) error {
	return resource.data.Delete(ctx, id)
}

func (resourceData *resourceData) Delete(ctx context.Context, id int64) error {
	q := resourceData.query(ctx).Is("id", id)
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

func (resource *Resource[T]) Name(singularName, pluralName func(string) string) *Resource[T] {
	resource.data.singularName = singularName
	resource.data.pluralName = pluralName
	return resource
}

func (resource *Resource[T]) PreviewURLFunction(fn func(*T) string) *Resource[T] {
	resource.data.PreviewURLFunction(func(a any) string {
		return fn(a.(*T))
	})
	return resource
}

func (resourceData *resourceData) PreviewURLFunction(fn func(any) string) *resourceData {
	resourceData.previewURLFunction = fn
	return resourceData
}

func (resource *Resource[T]) Icon(icon string) *Resource[T] {
	resource.data.icon = icon
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

func (resource *Resource[T]) Dashboard(name func(string) string) *Dashboard {
	return resource.data.resourceBoard.Dashboard(name)
}

func (resourceData *resourceData) addValidation(validation Validation) {
	resourceData.validations = append(resourceData.validations, validation)
}

func (resource *Resource[T]) DeleteValidation(validation Validation) *Resource[T] {

	resource.data.deleteValidations = append(resource.data.deleteValidations, validation)
	return resource
}

func (resource *Resource[T]) Board(board *Board) *Resource[T] {
	resource.data.parentBoard = board
	return resource
}

func (resourceData *resourceData) getItemURL(item interface{}, suffix string, userData UserData) string {
	ret := resourceData.getURL(fmt.Sprintf("%d", resourceData.previewer(userData, item).ID()))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func (app *App) getResourceByID(name string) *resourceData {
	return app.resourceNameMap[columnName(name)]
}

func initResource(resourceData *resourceData) {
	resourceData.resourceController.addAroundAction(func(request *Request, next func()) {
		if !request.Authorize(resourceData.canView) {
			renderErrorPage(request, 403)
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

func (resourceData *resourceData) getCachedCount(ctx context.Context) int64 {
	return loadCache(resourceData.app.cache, resourceData.cachedCountName(), func(ctx context.Context) int64 {
		count, _ := resourceData.countAllItems(ctx, false)
		return count
	})
}

func (resourceData *resourceData) updateCachedCount(ctx context.Context) error {
	resourceData.app.cache.forceLoad(resourceData.cachedCountName(), func(ctx context.Context) interface{} {
		count, _ := resourceData.countAllItems(ctx, false)
		return count
	})
	return nil
}
