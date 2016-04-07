package admin

import (
	"database/sql"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"math"
	"reflect"
	"strings"
	"time"
)

var Debug = false

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

func createTable(db *sql.DB, tableName string, adminStruct *AdminStructCache) error {
	description, err := adminStruct.getStructDescription()
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

func migrateTable(db *sql.DB, tableName string, adminStruct *AdminStructCache) error {
	structDescription, err := adminStruct.getStructDescription()
	if err != nil {
		return err
	}

	tableDescription, err := getTableDescription(db, tableName)
	if err != nil {
		return err
	}

	tableDescriptionMap := map[string]bool{}
	for _, item := range tableDescription {
		tableDescriptionMap[item.Field] = true
	}

	var columns []*mysqlColumn

	for _, item := range structDescription {
		if !tableDescriptionMap[item.Field] {
			columns = append(columns, item)
		}
	}

	if len(columns) == 0 {
		return nil
	}

	items := []string{}

	for _, v := range columns {
		additional := ""
		if v.Field == "id" {
			additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
		}
		item := fmt.Sprintf("ADD COLUMN %s %s %s", v.Field, v.Type, additional)
		items = append(items, item)
	}

	q := fmt.Sprintf("ALTER TABLE %s %s;", tableName, strings.Join(items, ", "))
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

func prepareValues(value reflect.Value) (names []string, questionMarks []string, values []interface{}, err error) {

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

type listQuery struct {
	whereString string
	whereParams []interface{}
	limit       int64
	offset      int64
	order       []listQueryOrder
}

type listQueryOrder struct {
	name string
	desc bool
}

func buildOrderString(params []listQueryOrder) string {
	if len(params) == 0 {
		params = []listQueryOrder{{name: "id"}}
	}
	items := []string{}
	for _, v := range params {
		order := "ASC"
		if v.desc {
			order = "DESC"
		}
		item := fmt.Sprintf("`%s` %s", v.name, order)
		items = append(items, item)
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(items, ", "))
}

func buildLimitString(offset, limit int64) string {
	if limit <= 0 {
		limit = math.MaxInt64
	}
	return fmt.Sprintf("LIMIT %d, %d", offset, limit)
}

func buildLimitWithoutOffsetString(limit int64) string {
	if limit <= 0 {
		limit = math.MaxInt64
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func buildWhereString(where string) string {
	if len(where) == 0 {
		where = "1"
	}
	return fmt.Sprintf("WHERE %s", where)
}

func mapToDBQuery(m map[string]interface{}) (str string, params []interface{}) {
	items := []string{}
	for k, v := range m {
		item := fmt.Sprintf("`%s`=?", k)
		items = append(items, item)
		params = append(params, v)
	}
	str = strings.Join(items, " AND ")
	return
}

func countItems(db *sql.DB, tableName string, query *listQuery) (int64, error) {
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.whereString)

	q := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s %s %s;", tableName, whereString, orderString, limitString)
	rows, err := db.Query(q, query.whereParams...)
	if err != nil {
		return -1, err
	}
	rows.Next()

	var i int64
	err = rows.Scan(&i)
	return i, err
}

func getFirstItem(structCache *AdminStructCache, db *sql.DB, tableName string, item interface{}, query *listQuery) error {
	var items interface{}
	err := listItems(structCache, db, tableName, &items, query)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(items)

	if val.Len() > 0 {
		reflect.ValueOf(item).Elem().Set(val.Index(0))
		return nil
	} else {
		return ErrorNotFound
	}
}

func listItems(structCache *AdminStructCache, db *sql.DB, tableName string, items interface{}, query *listQuery) error {
	slice := reflect.New(reflect.SliceOf(reflect.PtrTo(structCache.typ))).Elem()
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.whereString)

	newValue := reflect.New(structCache.typ).Elem()
	names, scanners, err := structCache.getStructScanners(newValue)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` %s %s %s;", strings.Join(names, ", "), tableName, whereString, orderString, limitString)
	if Debug {
		fmt.Println(q, query.whereParams)
	}
	rows, err := db.Query(q, query.whereParams...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		newValue = reflect.New(structCache.typ)
		names, scanners, err = structCache.getStructScanners(newValue.Elem())
		if err != nil {
			return err
		}
		rows.Scan(scanners...)
		slice.Set(reflect.Append(slice, newValue))
	}

	reflect.ValueOf(items).Elem().Set(slice)
	return nil
}

func deleteItems(db *sql.DB, tableName string, query *listQuery) (int64, error) {
	limitString := buildLimitWithoutOffsetString(query.limit)
	whereString := buildWhereString(query.whereString)

	q := fmt.Sprintf("DELETE FROM `%s` %s %s;", tableName, whereString, limitString)
	res, err := db.Exec(q, query.whereParams...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
