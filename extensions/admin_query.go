package extensions

import (
	"errors"
	"fmt"
	"reflect"
)

func (a *Admin) Create(item interface{}) error {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := a.resourceMap[typ]
	if !ok {
		return errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
	}

	return resource.Create(item)
}

type AdminQuery struct {
	query listQuery
	admin *Admin
	err   error
}

func (a *Admin) Query() *AdminQuery {
	return &AdminQuery{
		admin: a,
	}
}

func (aq *AdminQuery) Get(item interface{}) error {
	if aq.err != nil {
		return aq.err
	}

	var err error
	slice := false

	typ := reflect.TypeOf(item).Elem()

	if typ.Kind() == reflect.Slice {
		slice = true
		typ = typ.Elem().Elem()
	}

	resource, ok := aq.admin.resourceMap[typ]
	if !ok {
		return errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
	}

	var newItem interface{}
	if slice {
		err = listItems(resource.adminStructCache, aq.admin.db, resource.tableName(), &newItem, aq.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))
	} else {
		err = getFirstItem(resource.adminStructCache, aq.admin.db, resource.tableName(), &newItem, aq.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
	}
	return nil
}
