package prago

import (
	"database/sql"
	"fmt"
	"strings"
)

type Aggregation struct {
	listQuery    *listQuery
	aggregations []*aggregationItem
}

func (listQuery *listQuery) getAggregation() *Aggregation {
	return &Aggregation{
		listQuery: listQuery,
	}
}

func (q *QueryData[T]) Aggregation() *Aggregation {
	return q.listQuery.getAggregation()
}

type aggregationItem struct {
	FieldName string
	Typ       string
}

func (ai *aggregationItem) getRepresentation() string {
	if ai.Typ == "count" {
		return "count(*)"
	}
	return fmt.Sprintf("%s(%s)", ai.Typ, ai.FieldName)
}

func (agg *Aggregation) Sum(fieldName string) *Aggregation {
	agg.aggregations = append(agg.aggregations, &aggregationItem{
		FieldName: fieldName,
		Typ:       "sum",
	})
	return agg
}

func (agg *Aggregation) Count() *Aggregation {
	agg.aggregations = append(agg.aggregations, &aggregationItem{
		//FieldName: fieldName,
		Typ: "count",
	})
	return agg
}

func (agg *Aggregation) Min(fieldName string) *Aggregation {
	agg.aggregations = append(agg.aggregations, &aggregationItem{
		FieldName: fieldName,
		Typ:       "min",
	})
	return agg
}

func (agg *Aggregation) Max(fieldName string) *Aggregation {
	agg.aggregations = append(agg.aggregations, &aggregationItem{
		FieldName: fieldName,
		Typ:       "max",
	})
	return agg
}

func (agg *Aggregation) Get() ([]int64, error) {
	query := agg.listQuery
	orderString := buildOrderString(query.order)
	limitString := buildLimitString(query.offset, query.limit)
	whereString := buildWhereString(query.conditions)

	var scanners []any

	var aggFields []string
	for _, field := range agg.aggregations {
		aggFields = append(aggFields, field.getRepresentation())
		scanners = append(scanners, &intScanner{})
	}
	aggFieldsStr := strings.Join(aggFields, ", ")

	q := fmt.Sprintf("SELECT %s FROM `%s` %s %s %s;", aggFieldsStr, query.resource.id, whereString, orderString, limitString)
	if query.isDebug {
		fmt.Println(q, query.values)
	}

	rows, err := query.resource.app.db.QueryContext(query.context, q, query.values...)
	if err != nil {
		return nil, err
	}
	defer func() {
		rows.Close()
	}()
	rows.Next()

	err = rows.Scan(scanners...)
	if err != nil {
		return nil, err
	}

	var ret []int64
	for _, v := range scanners {
		intVal := v.(*intScanner).value
		ret = append(ret, intVal)
	}

	return ret, nil

}

type intScanner struct {
	value int64
}

func (s *intScanner) Scan(src interface{}) error {
	ni := sql.NullInt64{}
	err := ni.Scan(src)
	if err != nil {
		return err
	}
	s.value = ni.Int64
	return nil
}
