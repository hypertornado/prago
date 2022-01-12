package prago

import (
	"errors"
	"fmt"
	"reflect"
)

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

func (resource Resource[T]) Is(name string, value interface{}) *Query2[T] {
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
	count, err := resource.Query().query.Is("id", id).delete(&item)
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
