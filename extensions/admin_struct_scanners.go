package extensions

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
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

type scanner struct {
	value reflect.Value
}

func (s *scanner) Scan(src interface{}) error {
	var err error

	switch s.value.Type().Kind() {
	case reflect.Struct:
		nt := mysql.NullTime{}
		err := nt.Scan(src)
		if err != nil {
			return err
		}
		s.value.Set(reflect.ValueOf(nt.Time))
	case reflect.Bool:
		nb := sql.NullBool{}
		err := nb.Scan(src)
		if err != nil {
			return err
		}
		s.value.SetBool(nb.Bool)
	case reflect.String:
		ns := sql.NullString{}
		err = ns.Scan(src)
		if err != nil {
			return err
		}
		s.value.SetString(ns.String)
	case reflect.Int64:
		ni := sql.NullInt64{}
		err = ni.Scan(src)
		if err != nil {
			return err
		}
		s.value.SetInt(ni.Int64)
	}
	return nil
}