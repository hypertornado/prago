package extensions

import (
	"database/sql"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"reflect"
	"strings"
)

func dropTable(db *sql.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf("drop table `%s`;", tableName))
	return err
}

func createTable(db *sql.DB, tableName string, item interface{}) error {
	description, err := getStructDescription(item)
	if err != nil {
		return err
	}

	items := []string{}

	for _, v := range description {
		additional := ""
		if v.Field == "id" {
			additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
		}
		item := fmt.Sprintf("%s %s %s", v.Field, v.Type, additional)
		items = append(items, item)
	}

	q := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, strings.Join(items, ", "))
	_, err = db.Exec(q)
	return err
}

func getTableDescription(db *sql.DB, tableName string) (map[string]*mysqlColumn, error) {
	columns := map[string]*mysqlColumn{}
	rows, err := db.Query(fmt.Sprintf("describe `%s`;", tableName))
	if err != nil {
		return columns, err
	}
	defer rows.Close()

	for rows.Next() {
		column := &mysqlColumn{}
		rows.Scan(
			&column.Field,
			&column.Type,
			&column.Null,
			&column.Key,
			&column.Default,
			&column.Extra,
		)
		columns[column.Field] = column
	}

	return columns, nil
}

func getStructDescription(item interface{}) (map[string]*mysqlColumn, error) {
	columns := map[string]*mysqlColumn{}

	typ := reflect.TypeOf(item)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		use := true
		field := typ.Field(i)
		column := &mysqlColumn{
			Field: utils.PrettyUrl(field.Name),
		}

		switch field.Type.Kind().String() {
		case "int64":
			column.Type = "bigint(20)"
		case "string":
			column.Type = "varchar(255)"
		default:
			fmt.Println("Cant use field", field.Name)
			use = false
		}
		if use {
			columns[column.Field] = column
		}
	}
	return columns, nil
}

func getStructScanners(item interface{}) (map[string]*scanner, error) {
	typ := reflect.TypeOf(item).Elem()
	val := reflect.ValueOf(item).Elem()

	ret := map[string]*scanner{}

	for i := 0; i < typ.NumField(); i++ {
		use := true
		field := typ.Field(i)
		name := utils.PrettyUrl(field.Name)

		switch field.Type.Kind().String() {
		case "int64":
		case "string":
		default:
			fmt.Println("Cant use field", field.Name)
			use = false
		}
		if use {
			ret[name] = &scanner{val.Field(i)}
		}
	}

	return ret, nil
}

type scanner struct {
	value reflect.Value
}

func (s *scanner) Scan(src interface{}) error {
	var err error

	switch s.value.Type().Kind() {
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

func getItem(db *sql.DB, tableName string, item interface{}, id int64) error {
	scanners, err := getStructScanners(item)
	if err != nil {
		return err
	}

	keys := []string{}
	values := []interface{}{}
	for k, v := range scanners {
		keys = append(keys, k)
		values = append(values, v)
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` WHERE id=?", strings.Join(keys, ", "), tableName)
	rows, err := db.Query(q, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()

	return rows.Scan(values...)
}

func listItems(db *sql.DB, tableName string, items interface{}) error {
	return nil
}

func createItem(db *sql.DB, tableName string, item interface{}) error {
	value := reflect.ValueOf(item).Elem()

	names := []string{}
	questionMarks := []string{}
	values := []interface{}{}

	for i := 0; i < value.Type().NumField(); i++ {
		field := value.Type().Field(i)

		if field.Name == "ID" {
			continue
		}

		val := value.FieldByName(field.Name)

		switch field.Type.Kind() {
		case reflect.String:
			values = append(values, val.String())
		case reflect.Int64:
			values = append(values, val.Int())
		default:
			fmt.Println("wrong kind")
			continue
		}

		names = append(names, "`"+utils.PrettyUrl(field.Name)+"`")
		questionMarks = append(questionMarks, "?")
	}

	q := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", tableName, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	res, err := db.Exec(q, values...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	reflect.ValueOf(item).Elem().FieldByName("ID").SetInt(id)
	return nil
}
