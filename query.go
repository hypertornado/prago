package prago

import (
	"fmt"
	"reflect"
)

//Query represents query to db
type query struct {
	query *listQuery
	//app      *App
	err      error
	db       dbIface
	isDebug  bool
	resource *resource
}

func (app *App) create(item interface{}) error {
	resource, err := app.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.createWithDBIface(item, app.db, false)
}

func (app *App) update(item interface{}) error {
	resource, err := app.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.saveWithDBIface(item, app.db, false)
}

func (resource *resource) query() query {
	return query{
		query:    &listQuery{},
		db:       resource.app.db,
		resource: resource,
	}
}

func (q query) where(condition string, values ...interface{}) query {
	q.query.where(condition, values...)
	return q
}

func (q query) is(name string, value interface{}) query {
	return q.where(sqlFieldToQuery(name), value)
}

func (q query) debug() query {
	q.isDebug = true
	return q
}

func (q query) order(name string) query {
	q.query.addOrder(name, false)
	return q
}

func (q query) orderDesc(name string) query {
	q.query.addOrder(name, true)
	return q
}

func (q query) limit(limit int64) query {
	q.query.limit = limit
	return q
}

func (q query) offset(offset int64) query {
	q.query.offset = offset
	return q
}

func (q query) getOLD(item interface{}) error {
	if q.err != nil {
		return q.err
	}

	var err error
	slice := false

	typ := reflect.TypeOf(item).Elem()

	if typ.Kind() == reflect.Slice {
		slice = true
		typ = typ.Elem().Elem()
	}

	resource, ok := q.resource.app.resourceMap[typ]
	if !ok {
		return fmt.Errorf("can't find resource with type %s", typ)
	}

	var newItem interface{}
	if slice {
		err = listItems(*resource, q.db, resource.id, &newItem, q.query, q.isDebug)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))
	} else {
		err = getFirstItem(*resource, q.db, resource.id, &newItem, q.query, q.isDebug)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
	}
	return nil
}

func (q query) first() (interface{}, error) {
	var ret interface{}
	q.resource.newItem(&ret)
	err := q.getOLD(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (q query) list() (interface{}, error) {
	var items interface{}
	q.resource.newArrayOfItems(&items)
	err := q.getOLD(items)
	if err != nil {
		return nil, err
	}

	return reflect.ValueOf(items).Elem().Interface(), nil
	//return items, nil
}

func (q query) count() (int64, error) {
	/*resource, err := q.resource.app.getResourceByItem(item)
	if err != nil {
		return -1, err
	}*/
	return countItems(q.db, q.resource.id, q.query, q.isDebug)
}

func (q query) delete(item interface{}) (int64, error) {
	resource, err := q.resource.app.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return deleteItems(q.db, resource.id, q.query, q.isDebug)
}
