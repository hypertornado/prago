package prago

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

func (app *App) Create[T any](item *T) error {
	return app.CreateContext(context.Background(), item)
}

func (app *App) CreateContext[T any](ctx context.Context, item *T) error {
	resource := getResource[T](app)
	return resource.create(ctx, item)
}

func (resource *Resource) create(ctx context.Context, item any) error {
	resource.setTimestamp(item, "CreatedAt")
	resource.setTimestamp(item, "UpdatedAt")
	return resource.createItem(ctx, item, false)
}

func (request *Request) CreateWithLog[T any](item *T) error {
	resource := getResource[T](request.app)
	return resource.createWithLog(item, request)
}

func (resource *Resource) createWithLog(item any, userData UserData) error {
	err := resource.create(context.Background(), item)
	if err != nil {
		return err
	}
	err = resource.logActivity(userData, nil, item)
	if err != nil {
		return err
	}
	return resource.updateCachedCount()
}

func (app *App) Update[T any](item *T) error {
	resource := getResource[T](app)
	return resource.update(context.Background(), item, nil)
}

func (request *Request) UpdateWithLog[T any](item *T) error {
	resource := getResource[T](request.app)
	return resource.updateWithLog(item, request)
}

func (app *App) UpdatePartial[T any](item *T, fields []string) error {
	resource := getResource[T](app)
	onlyFields := map[string]bool{}
	for _, field := range fields {
		onlyFields[resource.Field(field).fieldClassName] = true
	}
	return resource.update(context.Background(), item, onlyFields)
}

func (resource *Resource) update(ctx context.Context, item any, onlyFields map[string]bool) error {
	resource.setTimestamp(item, "UpdatedAt")
	return resource.saveItem(ctx, item, onlyFields, false)
}

func (app *App) ReplaceContext[T any](ctx context.Context, item *T) error {
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

func (app *App) Delete[T any](id int64) error {
	return app.DeleteContext[T](context.Background(), id)
}

func (app *App) DeleteContext[T any](ctx context.Context, id int64) error {
	resource := getResource[T](app)
	return resource.delete(ctx, id)
}

func (request *Request) DeleteWithLog[T any](item *T) error {
	resource := getResource[T](request.app)
	return resource.deleteWithLog(item, request)
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
