package extensions

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
)

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
	str = strings.Join(items, ", ")
	return
}

func countItems(db *sql.DB, tableName string, query listQuery) (int64, error) {
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

func getFirstItem(structCache *AdminStructCache, db *sql.DB, tableName string, sliceItemType reflect.Type, item interface{}, query listQuery) error {
	var items interface{}
	err := listItems(structCache, db, tableName, sliceItemType, &items, query)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(items)

	if val.Len() > 0 {
		reflect.ValueOf(item).Elem().Set(val.Index(0))
	}
	return nil
}

func listItems(structCache *AdminStructCache, db *sql.DB, tableName string, sliceItemType reflect.Type, items interface{}, query listQuery) error {
	slice := reflect.New(reflect.SliceOf(reflect.PtrTo(sliceItemType))).Elem()
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.whereString)

	newValue := reflect.New(sliceItemType).Elem()
	names, scanners, err := structCache.getStructScanners(newValue)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` %s %s %s;", strings.Join(names, ", "), tableName, whereString, orderString, limitString)
	rows, err := db.Query(q, query.whereParams...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		newValue = reflect.New(sliceItemType)
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

func deleteItems(db *sql.DB, tableName string, query listQuery) (int64, error) {
	limitString := buildLimitWithoutOffsetString(query.limit)
	whereString := buildWhereString(query.whereString)

	q := fmt.Sprintf("DELETE FROM `%s` %s %s;", tableName, whereString, limitString)
	res, err := db.Exec(q, query.whereParams...)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}
