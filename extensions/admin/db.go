package admin

import (
	"errors"
	"fmt"
	"reflect"
)

//Create item in db
func (a *Admin) Create(item interface{}) error {
	resource, err := a.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.Create(item)
}

//Save item to db
func (a *Admin) Save(item interface{}) error {
	resource, err := a.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.Save(item)
}

//Query item from db
func (a *Admin) Query() *AdminQuery {
	return &AdminQuery{
		query: &listQuery{},
		admin: a,
		db:    a.getDB(),
	}
}

type AdminQuery struct {
	query *listQuery
	admin *Admin
	db    dbIface
	err   error
}

func (q *AdminQuery) Where(w ...interface{}) *AdminQuery {
	if q.err == nil {
		q.err = q.query.where(w...)
	}
	return q
}

func (q *AdminQuery) WhereIs(name string, value interface{}) *AdminQuery {
	return q.Where(map[string]interface{}{name: value})
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
		err = listItems(resource.StructCache, aq.db, resource.tableName(), &newItem, aq.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))
	} else {
		err = getFirstItem(resource.StructCache, aq.db, resource.tableName(), &newItem, aq.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
	}
	return nil
}

func (aq *AdminQuery) Count(item interface{}) (int64, error) {
	resource, err := aq.admin.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return countItems(aq.db, resource.tableName(), aq.query)
}

func (aq *AdminQuery) Delete(item interface{}) (int64, error) {
	resource, err := aq.admin.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return deleteItems(aq.db, resource.tableName(), aq.query)
}
