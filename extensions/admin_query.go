package extensions

import (
	"errors"
	"fmt"
	"reflect"
)

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

/*func (aq *AdminQuery) GetList(item interface{}) error {
	if aq.err != nil {
		return aq.err
	}

	typ := reflect.TypeOf(item).Elem()

	fmt.Println(typ.Kind())

}*/

func (aq *AdminQuery) Get(item interface{}) error {
	if aq.err != nil {
		return aq.err
	}

	typ := reflect.TypeOf(item).Elem()

	if typ.Kind() == reflect.Slice {
		typ = typ.Elem().Elem()

		resource, ok := aq.admin.resourceMap[typ]
		if !ok {
			return errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
		}

		var newItem interface{}
		err := listItems(resource.adminStructCache, aq.admin.db, resource.tableName(), &newItem, aq.query)
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))

		return err
	} else {
		resource, ok := aq.admin.resourceMap[typ]
		if !ok {
			return errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
		}

		var newItem interface{}
		err := getFirstItem(resource.adminStructCache, aq.admin.db, resource.tableName(), &newItem, aq.query)
		if err != nil {
			return err
		}

		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
		return nil
	}
}
