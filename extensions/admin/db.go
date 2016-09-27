package admin

import (
	"fmt"
	"reflect"
)

//Query represents query to db
type Query struct {
	query *listQuery
	admin *Admin
	db    dbIface
	err   error
}

//Create item in db
func (a *Admin) Create(item interface{}) error {
	resource, err := a.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.create(item)
}

//Save item to db
func (a *Admin) Save(item interface{}) error {
	resource, err := a.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.save(item)
}

//Query item from db
func (a *Admin) Query() *Query {
	return &Query{
		query: &listQuery{},
		admin: a,
		db:    a.getDB(),
	}
}

//Where adds where query
func (q *Query) Where(w ...interface{}) *Query {
	if q.err == nil {
		q.err = q.query.where(w...)
	}
	return q
}

//WhereIs adds where query for single item
func (q *Query) WhereIs(name string, value interface{}) *Query {
	return q.Where(map[string]interface{}{name: value})
}

//Order sets order column
func (q *Query) Order(name string) *Query {
	q.query.addOrder(name, false)
	return q
}

//OrderDesc sets descending order column
func (q *Query) OrderDesc(name string) *Query {
	q.query.addOrder(name, true)
	return q
}

//Limit query's result
func (q *Query) Limit(limit int64) *Query {
	q.query.limit = limit
	return q
}

//Offset of query's result
func (q *Query) Offset(offset int64) *Query {
	q.query.offset = offset
	return q
}

//Get item or items with query
func (q *Query) Get(item interface{}) error {
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

	resource, ok := q.admin.resourceMap[typ]
	if !ok {
		return fmt.Errorf("Can't find resource with type %s.", typ)
	}

	var newItem interface{}
	if slice {
		err = listItems(resource.StructCache, q.db, resource.tableName(), &newItem, q.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))
	} else {
		err = getFirstItem(resource.StructCache, q.db, resource.tableName(), &newItem, q.query)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
	}
	return nil
}

//Count items with query
func (q *Query) Count(item interface{}) (int64, error) {
	resource, err := q.admin.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return countItems(q.db, resource.tableName(), q.query)
}

//Delete item with query
func (q *Query) Delete(item interface{}) (int64, error) {
	resource, err := q.admin.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return deleteItems(q.db, resource.tableName(), q.query)
}
