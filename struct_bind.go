package prago

import (
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func (resource *Resource) setOrderPosition(item interface{}, order int64) error {
	value := reflect.ValueOf(item)

	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	val := value.FieldByName(resource.orderField.Name)
	val.SetInt(order)
	return nil
}

func (resource Resource) getDefaultBindedFieldIDs(user *user) map[string]bool {
	ret := map[string]bool{}
	for _, v := range resource.fieldArrays {
		if !v.authorizeView(user) {
			continue
		}
		if !v.authorizeEdit(user) {
			continue
		}
		ret[v.ColumnName] = true
	}
	return ret
}

func (resource Resource) bindData(item interface{}, user *user, params url.Values, bindedFieldIDs map[string]bool) error {

	if bindedFieldIDs == nil {
		bindedFieldIDs = resource.getDefaultBindedFieldIDs(user)
	}

	value := reflect.ValueOf(item)
	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	for _, field := range resource.fieldArrays {
		if !field.authorizeEdit(user) {
			continue
		}
		if !bindedFieldIDs[field.ColumnName] {
			continue
		}

		val := value.FieldByName(field.Name)
		urlValue := params.Get(field.ColumnName)

		switch field.Typ.Kind() {
		case reflect.Struct:
			if field.Typ == reflect.TypeOf(time.Now()) {
				if urlValue == "" {
					val.Set(reflect.ValueOf(time.Time{}))
				}
				if field.Tags["prago-type"] == "timestamp" {
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
