package pragelastic

import (
	"fmt"
	"reflect"
	"time"
)

type field struct {
	Name             string
	Type             string
	Analyzer         string
	CategoryContexts []SuggestCategoryContext
	Enabled          bool
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
	ret = &field{
		Enabled: true,
	}
	ret.Name = t.Name

	enabledStr := t.Tag.Get("elastic-enabled")
	if enabledStr == "false" {
		ret.Enabled = false
		ret.Type = "object"
		return
	}

	if ret.Name == "ID" {
		ret.Type = "keyword"
		return
	}

	if t.Type == reflect.TypeOf(time.Now()) {
		ret.Type = "date"
		return ret
	}

	/*if t.Type == reflect.TypeOf(map[string]float64{}) {
		ret.Type = "object"
		return ret
	}*/

	if t.Type == reflect.TypeOf([]string{}) {
		ret.Type = "keyword"
		return ret
	}
	if t.Type == reflect.TypeOf(Suggest{}) {
		ret.Type = "completion"
		ret.Analyzer = t.Tag.Get("elastic-analyzer")
		categoryContextName := t.Tag.Get("elastic-category-context-name")
		if categoryContextName != "" {
			ret.CategoryContexts = append(ret.CategoryContexts, SuggestCategoryContext{
				Name: categoryContextName,
				Type: "category",
			})
		}
		return ret
	}
	switch t.Type.Kind() {
	case reflect.String:
		typ := t.Tag.Get("elastic-datatype")
		if typ == "" {
			panic(fmt.Sprintf("string type '%s' must have elastic-datatype tag set", t.Name))
		}
		if !stringTypeValid(typ) {
			panic(fmt.Sprintf("string type '%s' have elastic-datatype tag: %s", t.Name, typ))
		}
		ret.Type = typ
		if ret.Type == "text" {
			ret.Analyzer = t.Tag.Get("elastic-analyzer")
		}
	case reflect.Map:
		//TODO: make supported maps more strict
		ret.Type = "object"
	case reflect.Int64:
		ret.Type = "long"
	case reflect.Float64:
		ret.Type = "double"
	case reflect.Bool:
		ret.Type = "boolean"
	default:
		panic("wrong type " + t.Type.Name())
	}
	return
}

func stringTypeValid(t string) bool {
	switch t {
	case "match_only_text":
		return true
	case "text":
		return true
	case "keyword":
		return true
	}
	return false
}
