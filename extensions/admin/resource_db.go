package admin

import (
	"database/sql"
	"errors"
	"reflect"
	"time"
)

var ErrorWrongWhereFormat = errors.New("Wrong Where Format")

type ResourceQuery struct {
	query       *listQuery
	db          *sql.DB
	tableName   string
	structCache *StructCache
	err         error
}

func (q *listQuery) where(data ...interface{}) error {
	var whereParams []interface{}
	var whereString string
	var err error

	if len(data) == 0 {
		return ErrorWrongWhereFormat
	}
	if len(data) == 1 {
		whereString, whereParams, err = q.whereSingle(data[0])
		if err != nil {
			return err
		}
	} else {
		first, ok := data[0].(string)
		if !ok {
			return ErrorWrongWhereFormat
		}
		whereString = first
		whereParams = data[1:len(data)]
	}

	if len(q.whereString) > 0 {
		q.whereString += " and "
	}
	q.whereString += whereString
	q.whereParams = append(q.whereParams, whereParams...)

	return nil
}

func (q *listQuery) whereSingle(data interface{}) (whereString string, whereParams []interface{}, err error) {
	switch data.(type) {
	case string:
		whereString = data.(string)
	case int64:
		whereString, whereParams = mapToDBQuery(map[string]interface{}{"id": data.(int64)})
	case int:
		whereString, whereParams = mapToDBQuery(map[string]interface{}{"id": data.(int)})
	case map[string]interface{}:
		whereString, whereParams = mapToDBQuery(data.(map[string]interface{}))
	default:
		err = ErrorWrongWhereFormat
	}
	return
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func (ar *AdminResource) Save(item interface{}) error {
	return ar.saveWithDBIface(item, ar.admin.DB())
}

func (ar *AdminResource) saveWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrorDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() && val.FieldByName(fn).CanSet() && val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}

	return ar.StructCache.saveItem(db, ar.tableName(), item)
}

func (ar *AdminResource) Create(item interface{}) error {
	return ar.createWithDBIface(item, ar.admin.DB())
}

func (ar *AdminResource) createWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrorDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	var t time.Time
	for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
			//TODO: create test for not seting value on non-zero times
			reflect.ValueOf(&t).Elem().Set(field)
			if t.IsZero() {
				field.Set(timeVal)
			}
		}
	}
	return ar.StructCache.createItem(db, ar.tableName(), item)
}

func (ar *AdminResource) Query() *ResourceQuery {
	var err error
	if !ar.HasModel {
		err = ErrorDontHaveModel
	}
	return &ResourceQuery{
		query:       &listQuery{},
		db:          ar.db(),
		tableName:   ar.tableName(),
		structCache: ar.StructCache,
		err:         err,
	}
}

func (ar *AdminResource) NewItem() (item interface{}, err error) {
	reflect.ValueOf(&item).Elem().Set(reflect.New(ar.Typ))
	return
}

func (q *ResourceQuery) Where(w interface{}) *ResourceQuery {
	if q.err == nil {
		q.err = q.query.where(w)
	}
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
	q.query.offset = i
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
