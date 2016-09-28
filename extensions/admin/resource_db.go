package admin

import (
	"reflect"
	"time"
)

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
