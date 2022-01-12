package prago

import (
	"errors"
	"fmt"
	"reflect"
)

type Resource2[T any] struct {
	Resource *Resource
}

func NewResource[T any](app *App) *Resource2[T] {
	var item T
	ret := &Resource2[T]{
		Resource: app.oldNewResource(item),
	}
	itemTyp := reflect.TypeOf(item)
	app.resource2Map[itemTyp] = ret
	return ret
}

func GetResource[T any](app *App) *Resource2[T] {
	var item T
	itemTyp := reflect.TypeOf(item)
	ret, ok := app.resource2Map[itemTyp]
	if !ok {
		return nil
	}
	return ret.(*Resource2[T])

}

func (resource Resource2[T]) Is(name string, value interface{}) *Query2[T] {
	return resource.Query().Is(name, value)
}

func (resource Resource2[T]) Create(item *T) error {
	return resource.Resource.app.create(item)
}

func (resource Resource2[T]) Update(item *T) error {
	return resource.Resource.app.Save(item)
}

func (resource Resource2[T]) GetItemWithID(id int64) *T {
	return resource.Query().Is("id", id).First()
}

func (resource Resource2[T]) Delete(id int64) error {
	var item T
	count, err := resource.Query().query.Is("id", id).Delete(&item)
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

func (resource Resource2[T]) Count() (int64, error) {
	return resource.Query().Count()
}

/*func (resource Resource2[T]) Items() []T {
	return nil
}*/
