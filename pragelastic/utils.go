package pragelastic

import (
	"reflect"
)

func getID[T any](item *T) string {
	val := reflect.ValueOf(*item)
	field := val.FieldByName("ID")
	return field.String()
}

func getFields[T any]() (ret []field) {
	var item T
	typ := reflect.TypeOf(item)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		ret = append(ret, field{
			Name: f.Name,
			Type: getElasticType(f),
		})
	}

	return
}

func getElasticType(t reflect.StructField) string {
	switch t.Type.Kind() {
	case reflect.String:
		return "text"
	case reflect.Int64:
		return "long"
	default:
		panic("wrong type " + t.Type.Name())
	}
}
