package extensions

import (
	"database/sql"
	"reflect"
	"time"
)

type ResourceQuery struct {
	query       *listQuery
	db          *sql.DB
	tableName   string
	structCache *AdminStructCache
	err         error
}

func (q *listQuery) where(w map[string]interface{}) {
	q.whereString, q.whereParams = mapToDBQuery(w)
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func (ar *AdminResource) Save(item interface{}) error {
	if !ar.hasModel {
		return ErrorDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() && val.FieldByName(fn).CanSet() && val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}

	return saveItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) Create(item interface{}) error {
	if !ar.hasModel {
		return ErrorDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
		if val.FieldByName(fieldName).IsValid() && val.FieldByName(fieldName).CanSet() && val.FieldByName(fieldName).Type() == timeVal.Type() {
			val.FieldByName(fieldName).Set(timeVal)
		}
	}
	return createItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) Query() *ResourceQuery {
	var err error
	if !ar.hasModel {
		err = ErrorDontHaveModel
	}
	return &ResourceQuery{
		query:       &listQuery{},
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
	q.query.where(w)
	return q
}

func (q *ResourceQuery) Order(name string) *ResourceQuery {
	q.query.addOrder(name, false)
	return q
}

func (q *ResourceQuery) OrderDesc(name string) *ResourceQuery {
	q.query.addOrder(name, true)
	return q
}

func (q *ResourceQuery) Limit(i int64) *ResourceQuery {
	q.query.limit = i
	return q
}

func (q *ResourceQuery) Offset(i int64) *ResourceQuery {
	q.query.limit = i
	return q
}

func (q *ResourceQuery) Count() (int64, error) {
	return countItems(q.db, q.tableName, q.query)
}

func (q *ResourceQuery) First() (item interface{}, err error) {
	if q.err != nil {
		return nil, q.err
	}
	err = getFirstItem(q.structCache, q.db, q.tableName, &item, q.query)
	return
}

func (q *ResourceQuery) List() (items interface{}, err error) {
	if q.err != nil {
		return nil, q.err
	}
	err = listItems(q.structCache, q.db, q.tableName, &items, q.query)
	return
}

func (q *ResourceQuery) Delete() (count int64, err error) {
	if q.err != nil {
		return -1, q.err
	}
	return deleteItems(q.db, q.tableName, q.query)
}
