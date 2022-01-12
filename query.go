package prago

import (
	"fmt"
	"reflect"
)

//Query represents query to db
type Query struct {
	query *listQuery
	app   *App
	err   error
	db    dbIface
	debug bool
}

//MustCreate
//Deprecated
func (app *App) mustCreate(item interface{}) {
	err := app.create(item)
	if err != nil {
		panic(fmt.Sprintf("can't create: %s", err))
	}
}

//Create item in db
//Deprecated
func (app *App) create(item interface{}) error {
	resource, err := app.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.createWithDBIface(item, app.db, false)
}

//Save item to db
func (app *App) save(item interface{}) error {
	resource, err := app.getResourceByItem(item)
	if err != nil {
		return err
	}
	return resource.saveWithDBIface(item, app.db, false)
}

//Query item from db
func (app *App) Query() Query {
	return Query{
		query: &listQuery{},
		app:   app,
		db:    app.db,
	}
}

//Where adds where query
func (q Query) Where(condition string, values ...interface{}) Query {
	q.query.where(condition, values...)
	return q
}

//Is adds where query for single item
func (app *App) Is(name string, value interface{}) Query {
	q := app.Query()
	return q.Is(name, value)
}

//Is adds where query for single item
func (q Query) Is(name string, value interface{}) Query {
	return q.Where(sqlFieldToQuery(name), value)
}

//WhereIs adds where query for single item
func (q Query) Debug() Query {
	q.debug = true
	return q
}

//Order sets order column
func (q Query) Order(name string) Query {
	q.query.addOrder(name, false)
	return q
}

//OrderDesc sets descending order column
func (q Query) OrderDesc(name string) Query {
	q.query.addOrder(name, true)
	return q
}

//Limit query's result
func (q Query) Limit(limit int64) Query {
	q.query.limit = limit
	return q
}

//Offset of query's result
func (q Query) Offset(offset int64) Query {
	q.query.offset = offset
	return q
}

//Get item or items with query
func (q Query) MustGet(item interface{}) {
	err := q.Get(item)
	if err != nil {
		panic(fmt.Sprintf("can't get: %s", err))
	}
}

//Get item or items with query
func (q Query) Get(item interface{}) error {
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

	resource, ok := q.app.resourceMap[typ]
	if !ok {
		return fmt.Errorf("can't find resource with type %s", typ)
	}

	var newItem interface{}
	if slice {
		err = listItems(*resource, q.db, resource.id, &newItem, q.query, q.debug)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem))
	} else {
		err = getFirstItem(*resource, q.db, resource.id, &newItem, q.query, q.debug)
		if err != nil {
			return err
		}
		reflect.ValueOf(item).Elem().Set(reflect.ValueOf(newItem).Elem())
	}
	return nil
}

//Count items with query
func (q Query) Count(item interface{}) (int64, error) {
	resource, err := q.app.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return countItems(q.db, resource.id, q.query, q.debug)
}

//Delete item with query
func (q Query) delete(item interface{}) (int64, error) {
	resource, err := q.app.getResourceByItem(item)
	if err != nil {
		return -1, err
	}
	return deleteItems(q.db, resource.id, q.query, q.debug)
}
