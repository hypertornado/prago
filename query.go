package prago

type Query[T any] struct {
	resource  *Resource[T]
	listQuery *listQuery
}

func (resource *Resource[T]) Query() *Query[T] {
	ret := &Query[T]{
		resource:  resource,
		listQuery: &listQuery{},
	}
	return ret
}

func (resource *Resource[T]) ID(id any) *T {
	return resource.Query().ID(id)
}

func (q *Query[T]) ID(id any) *T {
	return q.Where(sqlFieldToQuery("id"), id).First()
}

func (q *Query[T]) Is(name string, value interface{}) *Query[T] {
	return q.Where(sqlFieldToQuery(name), value)
}

func (q *Query[T]) Where(condition string, values ...interface{}) *Query[T] {
	q.listQuery.where(condition, values...)
	return q
}

func (q *Query[T]) Limit(limit int64) *Query[T] {
	q.listQuery.limit = limit
	return q
}

func (q *Query[T]) Offset(limit int64) *Query[T] {
	q.listQuery.offset = limit
	return q
}

func (q *Query[T]) Order(order string) *Query[T] {
	q.listQuery.addOrder(order, false)
	return q
}

func (q *Query[T]) Debug() *Query[T] {
	q.listQuery.isDebug = true
	return q
}

func (q *Query[T]) OrderDesc(order string) *Query[T] {
	q.listQuery.addOrder(order, true)
	return q
}

func (q *Query[T]) List() []*T {
	items, err := q.resource.data.listItems(q.listQuery, q.listQuery.isDebug)
	if err != nil {
		panic(err)
	}
	return items.([]*T)
}

func (q *Query[T]) First() *T {
	items := q.List()
	if len(items) > 0 {
		return items[0]
	}
	return nil
}

func (q *Query[T]) Count() (int64, error) {
	return q.resource.data.countItems(q.listQuery, q.listQuery.isDebug)
}
