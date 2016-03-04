package extensions

import (
	"errors"
	"reflect"
	"time"
)

func (s *AdminStructCache) getStructScanners(value reflect.Value) (names []string, scanners []interface{}, err error) {
	if value.Type() != s.typ {
		err = errors.New("Types dont match")
		return
	}

	for _, v := range s.fieldArrays {
		use := true
		switch v.typ.Kind() {
		case reflect.Int64:
		case reflect.Bool:
		case reflect.String:
		case reflect.Struct:
			if v.typ != reflect.TypeOf(time.Now()) {
				use = false
			}
		default:
			use = false
		}
		if use {
			names = append(names, v.lowercaseName)
			scanners = append(scanners, &scanner{value.Field(v.order)})
		}
	}
	return
}
