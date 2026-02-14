package prago

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"golang.org/x/net/context"
)

//Implement posgresql: https://gist.github.com/cpursley/c8fb81fe8a7e5df038158bdfe0f06dbb

// ErrWrongWhereFormat is returned when where query has a bad format
var ErrWrongWhereFormat = errors.New("wrong where format")

type mysqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   sql.NullString
}

type dbIface interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

func (resource *Resource) prepareValues(value reflect.Value) (names []string, questionMarks []string, values []interface{}, err error) {
	for _, field := range resource.fields {
		val := value.FieldByName(field.fieldClassName)

		if field.fieldClassName == "ID" {
			//TODO: is it really necessary to omit ids?
			if val.Int() <= 0 {
				continue
			}
		}

		switch field.typ.Kind() {
		case reflect.Struct:
			if field.typ == reflect.TypeOf(time.Now()) {
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

		names = append(names, "`"+columnName(field.id)+"`")
		questionMarks = append(questionMarks, "?")
	}
	return
}

// https://stackoverflow.com/questions/696190/create-if-an-entry-if-it-doesnt-exist-otherwise-update
func (resource *Resource) replaceItem(ctx context.Context, item interface{}, debugSQL bool) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	if id <= 0 {
		return errors.New("id must be positive")
	}
	value := reflect.ValueOf(item).Elem()
	names, questionMarks, values, err := resource.prepareValues(value)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES (%s);", resource.id, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	if debugSQL {
		fmt.Println(q, values)
	}
	execResult, err := resource.app.db.ExecContext(ctx, q, values...)
	if err != nil {
		return err
	}
	affected, err := execResult.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 && affected != 2 {
		return fmt.Errorf("not one or two rows affected: %d", affected)
	}
	return nil
}

func (resource *Resource) saveItem(ctx context.Context, item interface{}, debugSQL bool) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	if id <= 0 {
		return errors.New("id must be positive")
	}

	value := reflect.ValueOf(item).Elem()
	names, _, values, err := resource.prepareValues(value)
	if err != nil {
		return err
	}
	updateNames := []string{}
	for _, v := range names {
		updateNames = append(updateNames, fmt.Sprintf(" %s=? ", v))
	}
	q := fmt.Sprintf("UPDATE `%s` SET %s WHERE id=%d;", resource.id, strings.Join(updateNames, ", "), id)
	if debugSQL {
		fmt.Println(q, values)
	}
	execResult, err := resource.app.db.ExecContext(ctx, q, values...)
	if err != nil {
		return err
	}
	affected, err := execResult.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("zero rows affected by save operation")
	}
	if affected != 1 {
		return fmt.Errorf("non one row affected: %d", affected)
	}
	return nil
}

func (resource *Resource) createItem(ctx context.Context, item interface{}, debugSQL bool) error {
	value := reflect.ValueOf(item).Elem()

	names, questionMarks, values, err := resource.prepareValues(value)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", resource.id, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	if debugSQL {
		fmt.Println(q, values)
	}
	res, err := resource.app.db.ExecContext(ctx, q, values...)
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

func buildWhereString(conditions []string) string {
	var where string
	if len(conditions) == 0 {
		where = "1"
	} else {
		where = strings.Join(conditions, " AND ")
	}
	return fmt.Sprintf("WHERE %s", where)
}

func sqlFieldToQuery(fieldName string) string {
	return fmt.Sprintf("`%s`=?", fieldName)
}

func (resource *Resource) countAllItems() int64 {

	ret, err := resource.query(context.Background()).count()
	if err != nil {
		return -1
	}
	return ret

	/*var name string
	var rows sql.NullInt64

	query := fmt.Sprintf("SHOW TABLE STATUS LIKE '%s'", resource.id)
	row := resource.app.db.QueryRow(query)

	// You can scan only the columns you need
	err := row.Scan(
		&name,            // Name
		new(interface{}), // Engine
		new(interface{}), // Version
		new(interface{}), // Row_format
		&rows,            // Rows (this is what we want)
		new(interface{}), // Avg_row_length
		new(interface{}), // Data_length
		new(interface{}), // Max_data_length
		new(interface{}), // Index_length
		new(interface{}), // Data_free
		new(interface{}), // Auto_increment
		new(interface{}), // Create_time
		new(interface{}), // Update_time
		new(interface{}), // Check_time
		new(interface{}), // Collation
		new(interface{}), // Checksum
		new(interface{}), // Create_options
		new(interface{}), // Comment
	)
	if err != nil {
		resource.app.Log().Errorf("can't get resource stats (%s): %s", resource.id, err)
		return -1
	}

	return rows.Int64*/
}

func (app *App) getTableDataSize(tableName string) int64 {

	var name string
	var size sql.NullInt64

	query := fmt.Sprintf("SHOW TABLE STATUS LIKE '%s'", tableName)
	row := app.db.QueryRow(query)

	// You can scan only the columns you need
	err := row.Scan(
		&name,            // Name
		new(interface{}), // Engine
		new(interface{}), // Version
		new(interface{}), // Row_format
		new(interface{}), // Rows (this is what we want)
		new(interface{}), // Avg_row_length
		&size,            // Data_length
		new(interface{}), // Max_data_length
		new(interface{}), // Index_length
		new(interface{}), // Data_free
		new(interface{}), // Auto_increment
		new(interface{}), // Create_time
		new(interface{}), // Update_time
		new(interface{}), // Check_time
		new(interface{}), // Collation
		new(interface{}), // Checksum
		new(interface{}), // Create_options
		new(interface{}), // Comment
	)
	if err != nil {
		app.Log().Errorf("can't get resource stats (%s): %s", tableName, err)
		return -1
	}

	return size.Int64
}

func (query *listQuery) count() (int64, error) {
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.conditions)

	q := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s %s %s;", query.resource.id, whereString, orderString, limitString)
	if query.isDebug {
		fmt.Println(q, query.values)
	}

	rows, err := query.resource.app.db.QueryContext(query.context, q, query.values...)
	if err != nil {
		return -1, err
	}
	defer func() {
		rows.Close()
	}()
	rows.Next()

	var i int64
	err = rows.Scan(&i)
	return i, err
}

func (query *listQuery) list() (interface{}, error) {
	slice := reflect.New(reflect.SliceOf(reflect.PointerTo(query.resource.typ))).Elem()

	tableName := query.resource.id
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.conditions)

	names, _, err := query.resource.getStructScanners(
		reflect.New(query.resource.typ).Elem(),
	)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` %s %s %s;", strings.Join(names, ", "), tableName, whereString, orderString, limitString)
	if query.isDebug {
		query.resource.app.Log().Println(q, query.values)
	}

	rows, err := query.resource.app.db.QueryContext(query.context, q, query.values...)
	if err != nil {
		return nil, err
	}
	defer func() {
		rows.Close()
	}()
	for rows.Next() {
		newValue := reflect.New(query.resource.typ)

		_, scanners, err := query.resource.getStructScanners(newValue.Elem())
		if err != nil {
			return nil, err
		}
		rows.Scan(scanners...)
		slice.Set(reflect.Append(slice, newValue))
	}
	return slice.Interface(), nil
}

func (query *listQuery) delete() (int64, error) {
	resource := query.resource
	limitString := buildLimitWithoutOffsetString(query.limit)
	whereString := buildWhereString(query.conditions)

	q := fmt.Sprintf("DELETE FROM `%s` %s %s;", resource.id, whereString, limitString)
	if query.isDebug {
		resource.app.Log().Println(q, query.values)
	}
	res, err := resource.app.db.ExecContext(query.context, q, query.values...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
