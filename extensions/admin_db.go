package extensions

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago/utils"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
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

		switch field.Type.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if field.Type == dateType {
				column.Type = "datetime"
			} else {
				use = false
			}
		case reflect.Bool:
			column.Type = "bool"
		case reflect.Int64:
			column.Type = "bigint(20)"
		case reflect.String:
			if field.Tag.Get("prago-admin-type") == "text" {
				column.Type = "text"
			} else {
				column.Type = "varchar(255)"
			}
		default:
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

		switch field.Type.Kind() {
		case reflect.Int64:
		case reflect.Bool:
		case reflect.String:
		case reflect.Struct:
			if field.Type != reflect.TypeOf(time.Now()) {
				use = false
			}
		default:
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

func getItem(db *sql.DB, tableName string, itemType reflect.Type, item interface{}, id int64) error {
	whereString, whereParams := mapToDBQuery(map[string]interface{}{"id": id})
	return getFirstItem(db, tableName, itemType, item, listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
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
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Now()) {
				var tm time.Time
				reflect.ValueOf(&tm).Elem().Set(val)
				timeStr := tm.Format("2006-01-02 15:04:05")
				values = append(values, timeStr)
			} else {
				continue
			}
		case reflect.Bool:
			values = append(values, val.Bool())
		case reflect.String:
			values = append(values, val.String())
		case reflect.Int64:
			values = append(values, val.Int())
		default:
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

func deleteItem(db *sql.DB, tableName string, id int64) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE id=%d", tableName, id)
	_, err := db.Exec(q)
	return err
}

func BindDataFilterDefault(field reflect.StructField) bool {
	if field.Name == "ID" {
		return false
	}
	return true
}

func BindData(item interface{}, data url.Values, bindDataFilter func(reflect.StructField) bool) {
	value := reflect.ValueOf(item).Elem()
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
		default:
			continue
		}
	}
}
