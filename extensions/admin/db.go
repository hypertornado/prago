package admin

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
	query *listQuery
	admin *Admin
	err   error
}

func (a *Admin) Query() *AdminQuery {
	return &AdminQuery{
		query: &listQuery{},
		admin: a,
	}
}

func (q *AdminQuery) Where(w map[string]interface{}) *AdminQuery {
	q.query.where(w)
	return q
}

func (q *AdminQuery) Order(name string) *AdminQuery {
	q.query.addOrder(name, false)
	return q
}

func (q *AdminQuery) OrderDesc(name string) *AdminQuery {
	q.query.addOrder(name, true)
	return q
}

func (q *AdminQuery) Limit(i int64) *AdminQuery {
	q.query.limit = i
	return q
}

func (q *AdminQuery) Offset(i int64) *AdminQuery {
	q.query.offset = i
	return q
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

func (aq *AdminQuery) Count(item interface{}) (int64, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := aq.admin.resourceMap[typ]
	if !ok {
		return -1, errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
	}
	return countItems(aq.admin.db, resource.tableName(), aq.query)
}
