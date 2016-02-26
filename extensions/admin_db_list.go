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
	asc  bool
}

func buildOrderString(params []listQueryOrder) string {
	if len(params) == 0 {
		params = []listQueryOrder{{name: "id", asc: true}}
	}

	items := []string{}
	for _, v := range params {
		order := "ASC"
		if !v.asc {
			order = "DESC"
		}
		item := fmt.Sprintf("`%s` %s", v.name, order)
		items = append(items, item)
	}
	return strings.Join(items, ", ")
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

func listItems(db *sql.DB, tableName string, sliceItemType reflect.Type, items interface{}, query *listQuery) error {
	slice := reflect.New(reflect.SliceOf(sliceItemType)).Elem()

	if query == nil {
		query = &listQuery{}
	}

	orderString := buildOrderString(query.order)

	newValue := reflect.New(sliceItemType).Elem()
	names, scanners, err := getStructScanners(newValue)
	if err != nil {
		return err
	}

	var defaultLimit int64 = math.MaxInt64
	limit := defaultLimit
	if query.limit > 0 {
		limit = query.limit
	}
	limitString := fmt.Sprintf("LIMIT %d, %d", query.offset, limit)

	whereString := query.whereString
	if len(whereString) == 0 {
		whereString = "1"
	}

	q := fmt.Sprintf("SELECT %s FROM `%s` WHERE %s ORDER BY %s %s;", strings.Join(names, ", "), tableName, whereString, orderString, limitString)
	rows, err := db.Query(q, query.whereParams...)
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
