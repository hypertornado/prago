package admin

import (
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func (cache *StructCache) BindData(item interface{}, params url.Values, multiForm *multipart.Form, bindDataFilter StructFieldFilter) error {
	value := reflect.ValueOf(item)
	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	for _, field := range cache.fieldArrays {
		if !bindDataFilter(field) {
			continue
		}

		val := value.FieldByName(field.Name)

		urlValue := params.Get(field.Name)
		switch field.Typ.Kind() {
		case reflect.Struct:
			if field.Typ == reflect.TypeOf(time.Now()) {
				if field.Tags["prago-type"] == "timestamp" {
					tm, err := time.Parse("2006-01-02 15:04", urlValue)
					if err == nil {
						val.Set(reflect.ValueOf(tm))
					}
				} else {
					tm, err := time.Parse("2006-01-02", urlValue)
					if err == nil {
						val.Set(reflect.ValueOf(tm))
					}
				}
			}
		case reflect.String:
			if field.Tags["prago-type"] == "image" {
				imageId, err := NewImageFromMultipartForm(multiForm, field.Name)
				if err == nil {
					val.SetString(imageId)
				}
			} else {
				val.SetString(urlValue)
			}
		case reflect.Bool:
			if urlValue == "on" {
				val.SetBool(true)
			} else {
				val.SetBool(false)
			}
		case reflect.Int64:
			i, _ := strconv.Atoi(urlValue)
			val.SetInt(int64(i))
		}
	}
	return nil
}
