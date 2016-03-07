package extensions

import (
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func BindDataFilterDefault(field *adminStructField) bool {
	if field.name == "ID" {
		return false
	}
	return true
}

func (cache *AdminStructCache) BindDataNEW(item interface{}, params url.Values, form *multipart.Form, bindDataFilter func(reflect.StructField) bool) error {
	return nil
}

func (field *adminStructField) bindField(params url.Values, form *multipart.Form) {

}

func (cache *AdminStructCache) BindData(item interface{}, params url.Values, form *multipart.Form, bindDataFilter func(*adminStructField) bool) error {
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

		val := value.FieldByName(field.name)
		urlValue := params.Get(field.name)

		switch field.typ.Kind() {
		case reflect.Struct:
			if field.typ == reflect.TypeOf(time.Now()) {
				tm, err := time.Parse("2006-01-02", urlValue)
				if err == nil {
					val.Set(reflect.ValueOf(tm))
				}
			}
		case reflect.String:
			if field.tags["prago-admin-type"] == "image" {
				imageId, err := NewImageFromMultipartForm(form, field.name)
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
