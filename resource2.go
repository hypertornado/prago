package prago

import (
	"errors"
	"fmt"
	"reflect"
)

//https://github.com/golang/go/issues/49085

type Resource[T any] struct {
	id       string
	resource *resource
	app      *App

	previewURLFunction func(interface{}) string

	activityLog bool

	validations       []Validation
	deleteValidations []Validation

	resourceController *controller

	orderByColumn       string
	orderDesc           bool
	defaultItemsPerPage int64
}

func NewResource[T any](app *App) *Resource[T] {
	var item T

	defaultName := reflect.TypeOf(item).Name()

	ret := &Resource[T]{
		app: app,

		id: columnName(defaultName),

		activityLog: true,

		defaultItemsPerPage: 200,
		resourceController:  app.adminController.subController(),
	}
	itemTyp := reflect.TypeOf(item)
	app.resource2Map[itemTyp] = ret
	app.resources2 = append(app.resources2, ret)
	ret.resource = oldNewResource(ret, item)
	initResource(ret)

	ret.orderByColumn, ret.orderDesc = ret.getDefaultOrder()

	return ret
}

func GetResource[T any](app *App) *Resource[T] {
	var item T
	itemTyp := reflect.TypeOf(item)
	ret, ok := app.resource2Map[itemTyp]
	if !ok {
		return nil
	}
	return ret.(*Resource[T])

}

type resourceIface interface {
	initDefaultResourceActions()
	initDefaultResourceAPIs()
	initAutoRelations()
	addValidation(validation Validation)
	isOrderDesc() bool
	getOrderByColumn() string

	updateCachedCount() error
	getCachedCount() int64
	count() int64

	getResourceControl() *controller
	getID() string
}

func (resource Resource[T]) getID() string {
	return resource.id
}

func (resource Resource[T]) getResourceControl() *controller {
	return resource.resourceController
}

func (resource Resource[T]) isOrderDesc() bool {
	return resource.orderDesc
}

func (resource Resource[T]) getOrderByColumn() string {
	return resource.orderByColumn
}

func (resource Resource[T]) Is(name string, value interface{}) *Query[T] {
	return resource.Query().Is(name, value)
}

func (resource Resource[T]) Create(item *T) error {
	return resource.resource.app.create(item)
}

func (resource Resource[T]) Update(item *T) error {
	return resource.resource.app.update(item)
}

func (resource Resource[T]) Delete(id int64) error {
	var item T
	count, err := resource.Query().query.is("id", id).delete(&item)
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

func (resource Resource[T]) Count() (int64, error) {
	return resource.Query().Count()
}

func (resource *Resource[T]) Name(name func(string) string) *Resource[T] {
	resource.resource.name = name
	return resource
}

func (resource *Resource[T]) FieldName(nameOfField string, name func(string) string) *Resource[T] {
	resource.resource.FieldName(nameOfField, name)
	return resource
}

func (resource *Resource[T]) PreviewURLFunction(fn func(interface{}) string) *Resource[T] {
	resource.previewURLFunction = fn
	return resource
}

func (resource *Resource[T]) ItemsPerPage(itemsPerPage int64) *Resource[T] {
	resource.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *Resource[T]) PermissionView(permission Permission) *Resource[T] {
	resource.resource.PermissionView(permission)
	return resource
}

func (resource *Resource[T]) PermissionUpdate(permission Permission) *Resource[T] {
	resource.resource.PermissionUpdate(permission)
	return resource
}

func (resource *Resource[T]) PermissionCreate(permission Permission) *Resource[T] {
	resource.resource.PermissionCreate(permission)
	return resource
}

func (resource *Resource[T]) PermissionDelete(permission Permission) *Resource[T] {
	resource.resource.PermissionDelete(permission)
	return resource
}

func (resource *Resource[T]) PermissionExport(permission Permission) *Resource[T] {
	resource.resource.PermissionExport(permission)
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

func (resource *Resource[T]) FieldViewTemplate(IDofField string, viewTemplate string) *Resource[T] {
	resource.resource.FieldViewTemplate(IDofField, viewTemplate)
	return resource
}

func (resource *Resource[T]) FieldListCellTemplate(IDofField string, template string) *Resource[T] {
	resource.resource.FieldListCellTemplate(IDofField, template)
	return resource
}

func (resource *Resource[T]) FieldFormTemplate(IDofField string, template string) *Resource[T] {
	resource.resource.FieldFormTemplate(IDofField, template)
	return resource
}

func (resource *Resource[T]) FieldDBDescription(IDofField string, description string) *Resource[T] {
	resource.resource.FieldDBDescription(IDofField, description)
	return resource
}
