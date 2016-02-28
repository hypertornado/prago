package extensions

import (
	"database/sql"
	"reflect"
)

type adminResourceQuery struct {
	query         listQuery
	db            *sql.DB
	tableName     string
	sliceItemType reflect.Type
}

func (ar *AdminResource) Query() *adminResourceQuery {
	return &adminResourceQuery{
		query:         listQuery{},
		db:            ar.db(),
		tableName:     ar.tableName(),
		sliceItemType: ar.Typ,
	}
}

func (q *adminResourceQuery) Where(w map[string]interface{}) *adminResourceQuery {
	q.query.whereString, q.query.whereParams = mapToDBQuery(w)
	return q

}

func (q *adminResourceQuery) First() (item interface{}, err error) {
	err = getFirstItem(q.db, q.tableName, q.sliceItemType, &item, q.query)
	return
}

func (q *adminResourceQuery) List() (items interface{}, err error) {
	err = listItems(q.db, q.tableName, q.sliceItemType, &items, q.query)
	return
}
