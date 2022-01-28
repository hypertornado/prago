package pragelastic

import (
	"fmt"
	"reflect"
)

func getID[T any](item *T) string {
	val := reflect.ValueOf(*item)
	field := val.FieldByName("ID")
	return field.String()
}

func getFields[T any]() (ret []*field) {
	var item T
	typ := reflect.TypeOf(item)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		ret = append(ret, getElasticField(f))
	}

	return
}

func getElasticField(t reflect.StructField) (ret *field) {
	ret = &field{}
	ret.Name = t.Name
	if ret.Name == "ID" {
		ret.Type = "keyword"
		return
	}

	if t.Type == reflect.TypeOf([]string{}) {
		ret.Type = "keyword"
		return ret
	}
	if t.Type == reflect.TypeOf(Suggest{}) {
		ret.Type = "completion"
		ret.Analyzer = t.Tag.Get("elastic-analyzer")
		categoryContextName := t.Tag.Get("elastic-category-context-name")
		if categoryContextName != "" {
			//categoryContextPath := t.Tag.Get("elastic-category-context-path")
			ret.CategoryContexts = append(ret.CategoryContexts, SuggestCategoryContext{
				Name: categoryContextName,
				Type: "category",
				//Path: categoryContextPath,
			})
		}
		return ret
	}
	switch t.Type.Kind() {
	case reflect.String:
		typ := t.Tag.Get("elastic-datatype")
		if typ == "" {
			panic(fmt.Sprintf("string type '%s' must have elastic-datatype tag", t.Name))
		}
		if !stringTypeValid(typ) {
			panic(fmt.Sprintf("string type '%s' have elastic-datatype tag: %s", t.Name, typ))
		}
		ret.Type = typ
		if ret.Type == "text" {
			ret.Analyzer = t.Tag.Get("elastic-analyzer")
		}
	case reflect.Int64:
		ret.Type = "long"
	case reflect.Bool:
		ret.Type = "boolean"
	default:
		panic("wrong type " + t.Type.Name())
	}

	return
}

func stringTypeValid(t string) bool {
	switch t {
	case "text":
		return true
	case "keyword":
		return true
	}
	return false
}
