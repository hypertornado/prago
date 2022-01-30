package pragelastic

import (
	"reflect"
)

func getID[T any](item *T) string {
	val := reflect.ValueOf(*item)
	field := val.FieldByName("ID")
	return field.String()
}
