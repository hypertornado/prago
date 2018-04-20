package administration

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"time"
)

func (s *structCache) getStructScanners(value reflect.Value) (names []string, scanners []interface{}, err error) {
	if value.Type() != s.typ {
		err = errors.New("Types dont match")
		return
	}

	for _, v := range s.fieldArrays {
		use := true

		switch v.Typ.Kind() {
		case reflect.Int64:
		case reflect.Float64:
		case reflect.Bool:
		case reflect.String:
		case reflect.Struct:
			if v.Typ != reflect.TypeOf(time.Now()) {
				use = false
			}
		default:
			use = false
		}
		if use {
			names = append(names, v.ColumnName)
			scanners = append(scanners, &scanner{value.Field(v.fieldOrder)})
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
	case reflect.Float64:
		ni := sql.NullFloat64{}
		err = ni.Scan(src)
		if err != nil {
			return err
		}
		s.value.SetFloat(ni.Float64)
	}
	return nil
}
