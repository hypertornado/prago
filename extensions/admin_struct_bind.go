package extensions

import (
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func BindDataFilterDefault(field reflect.StructField) bool {
	if field.Name == "ID" {
		return false
	}
	return true
}

func (cache *AdminStructCache) BindData(item interface{}, params url.Values, form *multipart.Form, bindDataFilter func(reflect.StructField) bool) error {
	data := params

	value := reflect.ValueOf(item)
	for i := 0; i < 10; i++ {
		if value.Kind() == reflect.Struct {
			break
		}
		value = value.Elem()
	}

	for i := 0; i < value.Type().NumField(); i++ {
		field := value.Type().Field(i)

		if !bindDataFilter(field) {
			continue
		}

		val := value.FieldByName(field.Name)
		urlValue := data.Get(field.Name)

		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Now()) {
				tm, err := time.Parse("2006-01-02", urlValue)
				if err == nil {
					val.Set(reflect.ValueOf(tm))
				}
			}
		case reflect.String:
			if field.Tag.Get("prago-admin-type") == "image" {
				imageId, err := NewImageFromMultipartForm(form, field.Name)
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
		default:
			continue
		}
	}
	return nil
}
