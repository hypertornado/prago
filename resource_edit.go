package prago

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

func CreateItem[T any](app *App, item *T) error {
	return CreateItemWithContext(context.Background(), app, item)
}

func CreateItemWithContext[T any](ctx context.Context, app *App, item *T) error {
	resource := getResource[T](app)
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
	resource := getResource[T](app)
	return resource.update(ctx, item)
}

func (resource *Resource) update(ctx context.Context, item any) error {
	resource.setTimestamp(item, "UpdatedAt")
	return resource.saveItem(ctx, item, false)
}

func Replace[T any](ctx context.Context, app *App, item *T) error {
	resource := getResource[T](app)
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
	resource := getResource[T](app)
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
