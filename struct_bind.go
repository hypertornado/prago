package prago

import (
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func (resource *Resource) setOrderPosition(item interface{}, order int64) {
	value := reflect.ValueOf(item)

	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	val := value.FieldByName(resource.orderField.fieldClassName)
	val.SetInt(order)
}

func (resource *Resource) bindData(item any, request *Request, params url.Values) error {
	value := reflect.ValueOf(item)
	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	for _, field := range resource.fields {

		if !field.authorizeEdit(request) {
			continue
		}

		if len(params[field.id]) == 0 {
			continue
		}

		val := value.FieldByName(field.fieldClassName)
		urlValue := params.Get(field.id)

		if field.fixStringValueFN != nil {
			urlValue = field.fixStringValueFN(urlValue)
		}

		switch field.typ.Kind() {
		case reflect.Struct:
			if field.typ == reflect.TypeOf(time.Now()) {
				if urlValue == "" {
					val.Set(reflect.ValueOf(time.Time{}))
				}
				if field.tags["prago-type"] == "timestamp" {
					tm, err := time.Parse("2006-01-02 15:04", urlValue)
					if err == nil {
						val.Set(reflect.ValueOf(tm))
						continue
					}
				} else {
					tm, err := time.Parse("2006-01-02", urlValue)
					if err == nil {
						val.Set(reflect.ValueOf(tm))
						continue
					}
				}
			}
		case reflect.String:
			val.SetString(urlValue)
		case reflect.Bool:
			if urlValue == "on" {
				val.SetBool(true)
			} else {
				val.SetBool(false)
			}
		case reflect.Int64:
			i, _ := strconv.Atoi(urlValue)
			val.SetInt(int64(i))
		case reflect.Float64:
			i, _ := strconv.ParseFloat(urlValue, 64)
			val.SetFloat(i)
		}
	}
	return nil
}
