package prago

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
)

//https://github.com/golang/go/issues/49085

type Resource[T any] struct {
	id   string
	name func(locale string) string
	app  *App

	previewURLFunction func(*T) string

	activityLog bool

	validations       []Validation
	deleteValidations []Validation

	canView   Permission
	canCreate Permission
	canUpdate Permission
	canDelete Permission
	canExport Permission

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

	ret := &Resource[T]{
		app:  app,
		id:   columnName(defaultName),
		name: unlocalized(defaultName),

		canView:   sysadminPermission,
		canCreate: loggedPermission,
		canUpdate: loggedPermission,
		canDelete: loggedPermission,
		canExport: loggedPermission,

		typ: typ,

		activityLog: true,

		defaultItemsPerPage: 200,
		resourceController:  app.adminController.subController(),

		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := ret.newField(typ.Field(i), i)
			if field.tags["prago-type"] == "order" {
				ret.orderField = field
			}
			ret.fields = append(ret.fields, field)
			ret.fieldMap[field.columnName] = field
		}
	}

	app.resources = append(app.resources, ret)
	app.resourceMap[ret.typ] = ret
	app.resourceNameMap[ret.id] = ret

	initResource(ret)

	ret.orderByColumn, ret.orderDesc = ret.getDefaultOrder()
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
	createRelations()
	addValidation(validation Validation)
	isOrderDesc() bool
	getOrderByColumn() string

	updateCachedCount() error
	getCachedCount() int64
	count() int64

	addRelation(*relatedField)

	getResourceControl() *controller
	getID() string

	bindActions()

	getPermissionView() Permission
	getPermissionCreate() Permission
	getPermissionUpdate() Permission
	getPermissionDelete() Permission
	getPermissionExport() Permission

	getName(string) string
	getNameFunction() func(string) string

	getApp() *App

	unsafeDropTable() error
	migrate(bool) error

	getURL(suffix string) string

	//query() query

	getStructScanners(reflect.Value) ([]string, []interface{}, error)
	getTyp() reflect.Type

	getPreviewData(user *user, f *Field, value int64) (*preview, error)
	getnavigation2(action *Action, request *Request) navigation

	getPreviews(listRequest relationListRequest, user *user) []*preview
	importSearchData(e *adminSearch) error
	getItemPreview(id int64, user *user, relatedResource resourceIface) *preview
	resourceItemName(id int64) string
	itemWithRelationCount(fieldName string, id int64) int64
}

func (resource *Resource[T]) getTyp() reflect.Type {
	return resource.typ
}

func (resource *Resource[T]) getApp() *App {
	return resource.app
}

func (resource *Resource[T]) getNameFunction() func(string) string {
	return resource.name
}

func (resource *Resource[T]) getName(locale string) string {
	return resource.name(locale)
}

func (resource *Resource[T]) getPermissionView() Permission {
	return resource.canView
}

func (resource *Resource[T]) getPermissionCreate() Permission {
	return resource.canCreate
}

func (resource *Resource[T]) getPermissionUpdate() Permission {
	return resource.canUpdate
}

func (resource *Resource[T]) getPermissionDelete() Permission {
	return resource.canDelete
}

func (resource *Resource[T]) getPermissionExport() Permission {
	return resource.canDelete
}

func (resource *Resource[T]) addRelation(field *relatedField) {
	resource.relations = append(resource.relations, field)
}

func (resource *Resource[T]) getID() string {
	return resource.id
}

func (resource *Resource[T]) getResourceControl() *controller {
	return resource.resourceController
}

func (resource *Resource[T]) isOrderDesc() bool {
	return resource.orderDesc
}

func (resource *Resource[T]) getOrderByColumn() string {
	return resource.orderByColumn
}

func (resource *Resource[T]) Is(name string, value interface{}) *Query[T] {
	return resource.Query().Is(name, value)
}

func (resource *Resource[T]) Create(item *T) error {
	return resource.createWithDBIface(item, resource.app.db, false)
}

func (resource *Resource[T]) Update(item *T) error {
	return resource.saveWithDBIface(item, resource.app.db, false)
}

func (resource *Resource[T]) Delete(id int64) error {
	q := resource.Is("id", id)
	count, err := deleteItems(resource.app.db, resource.getID(), q.listQuery, q.isDebug)
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

func (resource *Resource[T]) Count() (int64, error) {
	return resource.Query().Count()
}

func (resource *Resource[T]) Name(name func(string) string) *Resource[T] {
	resource.name = name
	return resource
}

func (resource *Resource[T]) PreviewURLFunction(fn func(*T) string) *Resource[T] {
	resource.previewURLFunction = fn
	return resource
}

func (resource *Resource[T]) ItemsPerPage(itemsPerPage int64) *Resource[T] {
	resource.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *Resource[T]) PermissionView(permission Permission) *Resource[T] {
	must(resource.app.validatePermission(permission))
	resource.canView = permission
	return resource
}

func (resource *Resource[T]) PermissionUpdate(permission Permission) *Resource[T] {
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

func (resource *Resource[T]) PermissionCreate(permission Permission) *Resource[T] {
	must(resource.app.validatePermission(permission))
	resource.canCreate = permission
	return resource
}

func (resource *Resource[T]) PermissionDelete(permission Permission) *Resource[T] {
	must(resource.app.validatePermission(permission))
	resource.canDelete = permission
	return resource
}

func (resource *Resource[T]) PermissionExport(permission Permission) *Resource[T] {
	must(resource.app.validatePermission(permission))
	resource.canExport = permission
	return resource
}

func (resource *Resource[T]) Validation(validation Validation) *Resource[T] {
	resource.addValidation(validation)
	return resource
}

func (resource *Resource[T]) addValidation(validation Validation) {
	resource.validations = append(resource.validations, validation)
}

func (resource *Resource[T]) DeleteValidation(validation Validation) *Resource[T] {
	resource.deleteValidations = append(resource.deleteValidations, validation)
	return resource
}
