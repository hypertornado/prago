package prago

import (
	"errors"
	"fmt"
	"reflect"
)

//https://github.com/golang/go/issues/49085

type Resource[T any] struct {
	Resource *resource
}

func NewResource[T any](app *App) *Resource[T] {
	var item T
	ret := &Resource[T]{
		Resource: app.oldNewResource(item),
	}
	itemTyp := reflect.TypeOf(item)
	app.resource2Map[itemTyp] = ret
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

func (resource Resource[T]) Is(name string, value interface{}) *Query[T] {
	return resource.Query().Is(name, value)
}

func (resource Resource[T]) Create(item *T) error {
	return resource.Resource.app.create(item)
}

func (resource Resource[T]) Update(item *T) error {
	return resource.Resource.app.save(item)
}

func (resource Resource[T]) GetItemWithID(id int64) *T {
	return resource.Query().Is("id", id).First()
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
	resource.Resource.Name(name)
	return resource
}

func (resource *Resource[T]) PreviewURLFunction(fn func(interface{}) string) *Resource[T] {
	resource.Resource.PreviewURLFunction(fn)
	return resource
}

func (resource *Resource[T]) ItemsPerPage(itemsPerPage int64) *Resource[T] {
	resource.Resource.ItemsPerPage(itemsPerPage)
	return resource
}

func (resource *Resource[T]) PermissionView(permission Permission) *Resource[T] {
	resource.Resource.PermissionView(permission)
	return resource
}

func (resource *Resource[T]) PermissionEdit(permission Permission) *Resource[T] {
	resource.Resource.PermissionEdit(permission)
	return resource
}

func (resource *Resource[T]) PermissionCreate(permission Permission) *Resource[T] {
	resource.Resource.PermissionCreate(permission)
	return resource
}

func (resource *Resource[T]) PermissionDelete(permission Permission) *Resource[T] {
	resource.Resource.PermissionDelete(permission)
	return resource
}

func (resource *Resource[T]) PermissionExport(permission Permission) *Resource[T] {
	resource.Resource.PermissionExport(permission)
	return resource
}

func (resource *Resource[T]) Validation(validation Validation) *Resource[T] {
	resource.Resource.Validation(validation)
	return resource
}

func (resource *Resource[T]) DeleteValidation(validation Validation) *Resource[T] {
	resource.Resource.DeleteValidation(validation)
	return resource
}

func (resource *Resource[T]) FieldViewTemplate(IDofField string, viewTemplate string) *Resource[T] {
	resource.Resource.FieldViewTemplate(IDofField, viewTemplate)
	return resource
}

func (resource *Resource[T]) FieldListCellTemplate(IDofField string, template string) *Resource[T] {
	resource.Resource.FieldListCellTemplate(IDofField, template)
	return resource
}

func (resource *Resource[T]) FieldFormTemplate(IDofField string, template string) *Resource[T] {
	resource.Resource.FieldFormTemplate(IDofField, template)
	return resource
}

func (resource *Resource[T]) FieldDBDescription(IDofField string, description string) *Resource[T] {
	resource.Resource.FieldDBDescription(IDofField, description)
	return resource
}
