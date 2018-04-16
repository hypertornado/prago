package administration

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

//ErrWrongWhereFormat is returned when where query has a bad format
var ErrWrongWhereFormat = errors.New("wrong where format")

//Debug enables logging of all sql queries
var Debug = false

type mysqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   sql.NullString
}

type listQueryOrder struct {
	name string
	desc bool
}

type listQuery struct {
	whereString string
	whereParams []interface{}
	limit       int64
	offset      int64
	order       []listQueryOrder
}

type dbIface interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

func (q *listQuery) where(data ...interface{}) error {
	var whereParams []interface{}
	var whereString string
	var err error

	if len(data) == 0 {
		return ErrWrongWhereFormat
	}
	if len(data) == 1 {
		whereString, whereParams, err = q.whereSingle(data[0])
		if err != nil {
			return err
		}
	} else {
		first, ok := data[0].(string)
		if !ok {
			return ErrWrongWhereFormat
		}
		whereString = first
		whereParams = data[1:len(data)]
	}

	if len(q.whereString) > 0 {
		q.whereString += " AND "
	}
	q.whereString += whereString
	q.whereParams = append(q.whereParams, whereParams...)

	return nil
}

func (q *listQuery) whereSingle(data interface{}) (whereString string, whereParams []interface{}, err error) {
	switch data.(type) {
	case string:
		whereString = data.(string)
	case int64:
		whereString, whereParams = mapToDBQuery(map[string]interface{}{"id": data.(int64)})
	case int:
		whereString, whereParams = mapToDBQuery(map[string]interface{}{"id": data.(int)})
	case map[string]interface{}:
		whereString, whereParams = mapToDBQuery(data.(map[string]interface{}))
	default:
		err = ErrWrongWhereFormat
	}
	return
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func dropTable(db dbIface, tableName string) error {
	_, err := db.Exec(fmt.Sprintf("drop table `%s`;", tableName))
	return err
}

func listTables(db dbIface) (ret map[string]bool, err error) {
	ret = make(map[string]bool)
	var rows *sql.Rows
	rows, err = db.Query("show tables;")
	defer rows.Close()
	if err != nil {
		return ret, err
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return
		}
		ret[name] = true
	}
	return
}

func createTable(db dbIface, tableName string, adminStruct *structCache, verbose bool) (err error) {
	if verbose {
		fmt.Printf("Creating table '%s'\n", tableName)
	}
	items := []string{}
	for _, v := range adminStruct.fieldArrays {
		items = append(items, v.fieldDescriptionMysql(adminStruct.fieldTypes))
	}
	q := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, strings.Join(items, ", "))
	if verbose || Debug {
		fmt.Printf(" %s\n", q)
	}
	_, err = db.Exec(q)
	return err
}

func migrateTable(db dbIface, tableName string, adminStruct *structCache, verbose bool) error {
	if verbose {
		fmt.Printf("Migrating table '%s'\n", tableName)
	}
	tableDescription, err := getTableDescription(db, tableName)
	if err != nil {
		return err
	}

	tableDescriptionMap := map[string]bool{}
	for _, item := range tableDescription {
		tableDescriptionMap[item.Field] = true
	}

	items := []string{}

	for _, v := range adminStruct.fieldArrays {
		if !tableDescriptionMap[v.ColumnName] {
			items = append(items, fmt.Sprintf("ADD COLUMN %s", v.fieldDescriptionMysql(adminStruct.fieldTypes)))
		} else {
			tableDescriptionMap[v.ColumnName] = false
		}
	}

	if verbose {
		unusedFields := []string{}
		for k, v := range tableDescriptionMap {
			if v == true {
				unusedFields = append(unusedFields, k)
			}
		}
		if len(unusedFields) > 0 {
			fmt.Printf(" unused fields in model: %s\n", strings.Join(unusedFields, ", "))
		}
	}

	if len(items) == 0 {
		return nil
	}

	q := fmt.Sprintf("ALTER TABLE %s %s;", tableName, strings.Join(items, ", "))
	if verbose || Debug {
		fmt.Printf(" %s\n", q)
	}
	_, err = db.Exec(q)

	return err
}

func getTableDescription(db dbIface, tableName string) (map[string]*mysqlColumn, error) {
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

func (sc *structCache) prepareValues(value reflect.Value) (names []string, questionMarks []string, values []interface{}, err error) {

	for _, field := range sc.fieldArrays {
		val := value.FieldByName(field.Name)

		if field.Name == "ID" {
			//TODO: is it really necessary to omit ids?
			if val.Int() <= 0 {
				continue
			}
		}

		switch field.Typ.Kind() {
		case reflect.Struct:
			if field.Typ == reflect.TypeOf(time.Now()) {
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
		case reflect.Float64:
			values = append(values, val.Float())
		default:
			continue
		}

		names = append(names, "`"+columnName(field.Name)+"`")
		questionMarks = append(questionMarks, "?")
	}
	return
}

func (sc *structCache) saveItem(db dbIface, tableName string, item interface{}) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	value := reflect.ValueOf(item).Elem()
	names, _, values, err := sc.prepareValues(value)
	if err != nil {
		return err
	}
	updateNames := []string{}
	for _, v := range names {
		updateNames = append(updateNames, fmt.Sprintf(" %s=? ", v))
	}
	q := fmt.Sprintf("UPDATE `%s` SET %s WHERE id=%d;", tableName, strings.Join(updateNames, ", "), id)
	if Debug {
		fmt.Println(q, values)
	}
	execResult, err := db.Exec(q, values...)
	if err != nil {
		return err
	}
	affected, err := execResult.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("Zero rows affected by save operation.")
	}
	return nil
}

func (sc *structCache) createItem(db dbIface, tableName string, item interface{}) error {
	value := reflect.ValueOf(item).Elem()

	names, questionMarks, values, err := sc.prepareValues(value)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", tableName, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	if Debug {
		fmt.Println(q, values)
	}
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

func countItems(db dbIface, tableName string, query *listQuery) (int64, error) {
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.whereString)

	q := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s %s %s;", tableName, whereString, orderString, limitString)
	if Debug {
		fmt.Println(q, query.whereParams)
	}
	rows, err := db.Query(q, query.whereParams...)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	rows.Next()

	var i int64
	err = rows.Scan(&i)
	return i, err
}

func getFirstItem(cache *structCache, db dbIface, tableName string, item interface{}, query *listQuery) error {
	var items interface{}
	err := listItems(cache, db, tableName, &items, query)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(items)

	if val.Len() > 0 {
		reflect.ValueOf(item).Elem().Set(val.Index(0))
		return nil
	}
	return ErrItemNotFound
}

func listItems(cache *structCache, db dbIface, tableName string, items interface{}, query *listQuery) error {
	slice := reflect.New(reflect.SliceOf(reflect.PtrTo(cache.typ))).Elem()
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.whereString)

	newValue := reflect.New(cache.typ).Elem()
	names, scanners, err := cache.getStructScanners(newValue)
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
		newValue = reflect.New(cache.typ)
		names, scanners, err = cache.getStructScanners(newValue.Elem())
		if err != nil {
			return err
		}
		rows.Scan(scanners...)
		slice.Set(reflect.Append(slice, newValue))
	}

	reflect.ValueOf(items).Elem().Set(slice)
	return nil
}

func deleteItems(db dbIface, tableName string, query *listQuery) (int64, error) {
	limitString := buildLimitWithoutOffsetString(query.limit)
	whereString := buildWhereString(query.whereString)

	q := fmt.Sprintf("DELETE FROM `%s` %s %s;", tableName, whereString, limitString)
	if Debug {
		fmt.Println(q, query.whereParams)
	}
	res, err := db.Exec(q, query.whereParams...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}