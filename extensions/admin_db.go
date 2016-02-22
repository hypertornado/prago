package extensions

import (
	"database/sql"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"net/url"
	"reflect"
	"strings"
)

type mysqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   sql.NullString
}

func dropTable(db *sql.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf("drop table `%s`;", tableName))
	return err
}

func createTable(db *sql.DB, tableName string, typ reflect.Type) error {
	description, err := getStructDescription(typ)
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

func getStructDescription(typ reflect.Type) (map[string]*mysqlColumn, error) {
	columns := map[string]*mysqlColumn{}

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

func getStructScanners(value reflect.Value) (names []string, scanners []interface{}, err error) {
	names = []string{}
	scanners = []interface{}{}

	for i := 0; i < value.Type().NumField(); i++ {
		use := true
		field := value.Type().Field(i)
		name := utils.PrettyUrl(field.Name)

		switch field.Type.Kind().String() {
		case "int64":
		case "string":
		default:
			fmt.Println("Cant use field", field.Name)
			use = false
		}
		if use {
			names = append(names, name)
			scanners = append(scanners, &scanner{value.Field(i)})
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

func getItem(db *sql.DB, tableName string, itemType reflect.Type, item interface{}, id int64) error {
	value := reflect.New(itemType).Elem()
	names, scanners, err := getStructScanners(value)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` WHERE id=?", strings.Join(names, ", "), tableName)
	rows, err := db.Query(q, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()

	err = rows.Scan(scanners...)
	if err != nil {
		return err
	}

	reflect.ValueOf(item).Elem().Set(value)

	return nil
}

func listItems(db *sql.DB, tableName string, sliceItemType reflect.Type, items interface{}) error {
	slice := reflect.New(reflect.SliceOf(sliceItemType)).Elem()

	newValue := reflect.New(sliceItemType).Elem()
	names, scanners, err := getStructScanners(newValue)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s`;", strings.Join(names, ", "), tableName)
	rows, err := db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		newValue = reflect.New(sliceItemType).Elem()
		names, scanners, err = getStructScanners(newValue)
		if err != nil {
			return err
		}

		rows.Scan(scanners...)
		slice.Set(reflect.Append(slice, newValue))
	}

	reflect.ValueOf(items).Elem().Set(slice)
	return nil
}

func prepareValues(value reflect.Value) (names []string, questionMarks []string, values []interface{}, err error) {
	names = []string{}
	questionMarks = []string{}
	values = []interface{}{}

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
	return
}

func saveItem(db *sql.DB, tableName string, item interface{}) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	value := reflect.ValueOf(item).Elem()
	names, _, values, err := prepareValues(value)
	if err != nil {
		return err
	}
	updateNames := []string{}
	for _, v := range names {
		updateNames = append(updateNames, fmt.Sprintf(" %s=? ", v))
	}
	q := fmt.Sprintf("UPDATE `%s` SET %s WHERE id=%d;", tableName, strings.Join(updateNames, ", "), id)
	_, err = db.Exec(q, values...)
	return err
}

func createItem(db *sql.DB, tableName string, item interface{}) error {
	value := reflect.ValueOf(item).Elem()

	names, questionMarks, values, err := prepareValues(value)
	if err != nil {
		return err
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

func bindData(item interface{}, data url.Values) {
	value := reflect.ValueOf(item).Elem()
	for i := 0; i < value.Type().NumField(); i++ {
		field := value.Type().Field(i)

		if field.Name == "ID" {
			continue
		}

		val := value.FieldByName(field.Name)

		switch field.Type.Kind() {
		case reflect.String:
			val.SetString(data.Get(field.Name))
		case reflect.Int64:
			//values = append(values, val.Int())
		default:
			continue
		}
	}
}