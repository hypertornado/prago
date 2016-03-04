package extensions

import (
	"database/sql"
	"reflect"
)

type ResourceQuery struct {
	query       listQuery
	db          *sql.DB
	tableName   string
	structCache *AdminStructCache
	err         error
}

func (ar *AdminResource) Save(item interface{}) error {
	if !ar.hasModel {
		return ErrorDontHaveModel
	}
	return saveItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) Create(item interface{}) error {
	if !ar.hasModel {
		return ErrorDontHaveModel
	}
	return createItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) Query() *ResourceQuery {
	var err error
	if !ar.hasModel {
		err = ErrorDontHaveModel
	}
	return &ResourceQuery{
		query:       listQuery{},
		db:          ar.db(),
		tableName:   ar.tableName(),
		structCache: ar.adminStructCache,
		err:         err,
	}
}

func (ar *AdminResource) NewItem() (item interface{}, err error) {
	reflect.ValueOf(&item).Elem().Set(reflect.New(ar.Typ))
	return
}

func (q *ResourceQuery) Where(w map[string]interface{}) *ResourceQuery {
	q.query.whereString, q.query.whereParams = mapToDBQuery(w)
	return q
}

func (q *ResourceQuery) Order(name string) *ResourceQuery {
	q.query.order = append(q.query.order, listQueryOrder{name: name})
	return q
}

func (q *ResourceQuery) OrderDesc(name string) *ResourceQuery {
	q.query.order = append(q.query.order, listQueryOrder{name: name, desc: true})
	return q
}

func (q *ResourceQuery) Count() (int64, error) {
	return countItems(q.db, q.tableName, q.query)
}

func (q *ResourceQuery) First() (item interface{}, err error) {
	if q.err != nil {
		return nil, q.err
	}
	err = getFirstItem(q.structCache, q.db, q.tableName, q.structCache.typ, &item, q.query)
	return
}

func (q *ResourceQuery) List() (items interface{}, err error) {
	if q.err != nil {
		return nil, q.err
	}
	err = listItems(q.structCache, q.db, q.tableName, q.structCache.typ, &items, q.query)
	return
}

func (q *ResourceQuery) Delete() (count int64, err error) {
	if q.err != nil {
		return -1, q.err
	}
	return deleteItems(q.db, q.tableName, q.query)
}
