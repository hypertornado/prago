package prago

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type listQueryOrder struct {
	name string
	desc bool
}

type listQuery struct {
	context    context.Context
	conditions []string
	values     []interface{}
	limit      int64
	offset     int64
	order      []listQueryOrder
	isDebug    bool

	resource *Resource
}

func (resource *Resource) query(ctx context.Context) *listQuery {
	return &listQuery{
		context:  ctx,
		resource: resource,
	}
}

func (q *listQuery) where(condition string, values ...interface{}) *listQuery {
	q.conditions = append(q.conditions, condition)
	q.values = append(q.values, values...)
	return q
}

func inValueToArr(in any) (ret []any) {
	inStr, ok := in.(string)
	if ok {
		if inStr == "" {
			return nil
		}
		arr := strings.Split(inStr, ";")
		for _, v := range arr {
			if v == "" {
				continue
			}
			i, err := strconv.Atoi(v)
			if err != nil {
				panic("wrong number format")
			}
			ret = append(ret, int64(i))
		}
		return
	}
	inStrArr, ok := in.([]string)
	if ok {
		for _, v := range inStrArr {
			ret = append(ret, v)
		}
		return
	}
	inIntArr, ok := in.([]int64)
	if ok {
		for _, v := range inIntArr {
			ret = append(ret, v)
		}
		return
	}

	inInt64, ok := in.(int64)
	if ok {
		ret = append(ret, inInt64)
		return
	}

	panic("unknown option for multirelation")

}

func (q *listQuery) In(field string, value any) *listQuery {

	values := inValueToArr(value)
	if len(values) == 0 {
		return q
	}

	var placeholders []string
	for range values {
		placeholders = append(placeholders, "?")
	}

	q.where(fmt.Sprintf("`%s` IN (%s)", field, strings.Join(placeholders, ",")), values...)
	return q
}

func (q *listQuery) addOrder(name string, desc bool) {
	q.order = append(q.order, listQueryOrder{name: name, desc: desc})
}

func (listQuery *listQuery) Is(name string, value interface{}) *listQuery {
	listQuery.where(sqlFieldToQuery(name), value)
	return listQuery
}

func (listQuery *listQuery) ID(id any) any {
	listQuery.where(sqlFieldToQuery("id"), id)
	return listQuery.First()
}

func (listQuery *listQuery) First() any {
	items, err := listQuery.list()
	must(err)
	if reflect.ValueOf(items).Len() == 0 {
		return nil
	}

	return reflect.ValueOf(items).Index(0).Interface()
}

func (listQuery *listQuery) Limit(limit int64) *listQuery {
	listQuery.limit = limit
	return listQuery
}

func (listQuery *listQuery) Offset(offset int64) *listQuery {
	listQuery.offset = offset
	return listQuery
}

func (listQuery *listQuery) Order(order string) *listQuery {
	listQuery.addOrder(order, false)
	return listQuery
}

func (listQuery *listQuery) OrderDesc(order string) *listQuery {
	listQuery.addOrder(order, true)
	return listQuery
}
