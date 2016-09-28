package admin

import (
	"errors"
	"reflect"
	"time"
)

//ErrWrongWhereFormat is returned when where query has a bad format
var ErrWrongWhereFormat = errors.New("wrong where format")

func (q *listQuery) where(data ...interface{}) error {
	var whereParams []interface{}
	var whereString string
	var err error

	if len(data) == 0 {
		return ErrWrongWhereFormat
	}
	if len(data) == 1 {
		whereString, whereParams, err = q.whereSingle(data[0])
		if err != nil {
			return err
		}
	} else {
		first, ok := data[0].(string)
		if !ok {
			return ErrWrongWhereFormat
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
		err = ErrWrongWhereFormat
	}
	return
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func (ar *Resource) saveWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() && val.FieldByName(fn).CanSet() && val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}

	return ar.StructCache.saveItem(db, ar.tableName(), item)
}

func (ar *Resource) createWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	var t time.Time
	for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
			reflect.ValueOf(&t).Elem().Set(field)
			if t.IsZero() {
				field.Set(timeVal)
			}
		}
	}
	return ar.StructCache.createItem(db, ar.tableName(), item)
}

func (ar *Resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(ar.Typ))
}

func (ar *Resource) newItems(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(reflect.SliceOf(reflect.PtrTo(ar.Typ))))
}
