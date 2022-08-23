package prago

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

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

type listQueryOrder struct {
	name string
	desc bool
}

type listQuery struct {
	conditions []string
	values     []interface{}
	limit      int64
	offset     int64
	order      []listQueryOrder
}

type dbIface interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

func (q *listQuery) where(condition string, values ...interface{}) {
	q.conditions = append(q.conditions, condition)
	q.values = append(q.values, values...)
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func (resourceData *resourceData) prepareValues(value reflect.Value) (names []string, questionMarks []string, values []interface{}, err error) {
	for _, field := range resourceData.fields {
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
func (resourceData *resourceData) replaceItem(item interface{}, debugSQL bool) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	if id <= 0 {
		return errors.New("id must be positive")
	}

	value := reflect.ValueOf(item).Elem()
	names, questionMarks, values, err := resourceData.prepareValues(value)
	if err != nil {
		return err
	}
	updateNames := []string{}
	for _, v := range names {
		updateNames = append(updateNames, fmt.Sprintf(" %s=? ", v))
	}
	q := fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES (%s);", resourceData.id, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	if debugSQL {
		fmt.Println(q, values)
	}
	execResult, err := resourceData.app.db.Exec(q, values...)
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

func (resourceData *resourceData) saveItem(item interface{}, debugSQL bool) error {
	id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
	if id <= 0 {
		return errors.New("id must be positive")
	}

	value := reflect.ValueOf(item).Elem()
	names, _, values, err := resourceData.prepareValues(value)
	if err != nil {
		return err
	}
	updateNames := []string{}
	for _, v := range names {
		updateNames = append(updateNames, fmt.Sprintf(" %s=? ", v))
	}
	q := fmt.Sprintf("UPDATE `%s` SET %s WHERE id=%d;", resourceData.id, strings.Join(updateNames, ", "), id)
	if debugSQL {
		fmt.Println(q, values)
	}
	execResult, err := resourceData.app.db.Exec(q, values...)
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

func (resourceData *resourceData) createItem(item interface{}, debugSQL bool) error {
	value := reflect.ValueOf(item).Elem()

	names, questionMarks, values, err := resourceData.prepareValues(value)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", resourceData.id, strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	if debugSQL {
		fmt.Println(q, values)
	}
	res, err := resourceData.app.db.Exec(q, values...)
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

func (resourceData *resourceData) countAllItems(debugSQL bool) (int64, error) {
	return resourceData.countItems(&listQuery{}, debugSQL)
}

func (resourceData *resourceData) countItems(query *listQuery, debugSQL bool) (int64, error) {
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.conditions)

	q := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s %s %s;", resourceData.id, whereString, orderString, limitString)
	if debugSQL {
		fmt.Println(q, query.values)
	}
	rows, err := resourceData.app.db.Query(q, query.values...)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		return -1, err
	}
	rows.Next()

	var i int64
	err = rows.Scan(&i)
	return i, err
}

func (resource *Resource[T]) listItems(query *listQuery, debugSQL bool) ([]*T, error) {
	db := resource.data.app.db
	tableName := resource.data.id
	var ret []*T
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.conditions)

	newValue := reflect.New(resource.data.typ).Elem()
	names, _, err := resource.getStructScanners(newValue)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` %s %s %s;", strings.Join(names, ", "), tableName, whereString, orderString, limitString)
	if debugSQL {
		resource.data.app.Log().Println(q, query.values)
	}
	rows, err := db.Query(q, query.values...)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item T
		var ptrItem *T = &item
		newValue = reflect.ValueOf(ptrItem)
		_, scanners, err := resource.getStructScanners(newValue.Elem())
		if err != nil {
			return nil, err
		}
		rows.Scan(scanners...)
		ret = append(ret, ptrItem)
	}

	return ret, nil
}

func (resourceData *resourceData) deleteItems(query *listQuery, debugSQL bool) (int64, error) {

	limitString := buildLimitWithoutOffsetString(query.limit)
	whereString := buildWhereString(query.conditions)

	q := fmt.Sprintf("DELETE FROM `%s` %s %s;", resourceData.id, whereString, limitString)
	if debugSQL {
		resourceData.app.Log().Println(q, query.values)
	}
	res, err := resourceData.app.db.Exec(q, query.values...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
