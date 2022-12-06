package prago

import (
	"context"
	"reflect"
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

	resourceData *resourceData
}

func (resourceData *resourceData) query(ctx context.Context) *listQuery {
	return &listQuery{
		context:      ctx,
		resourceData: resourceData,
	}
}

/*func (resourceData *resourceData) Is(ctx context.Context, name string, value interface{}) *listQuery {
	listQuery := resourceData.query(ctx)
	listQuery.where(sqlFieldToQuery(name), value)
	return listQuery
}*/

/*func (resourceData *resourceData) ID(ctx context.Context, id any) any {
	return resourceData.query(ctx).ID(id)
}*/

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
