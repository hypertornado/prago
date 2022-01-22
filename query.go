package prago

type Query[T any] struct {
	resource  *Resource[T]
	listQuery *listQuery
	isDebug   bool
}

func (resource *Resource[T]) Query() *Query[T] {
	ret := &Query[T]{
		resource:  resource,
		listQuery: &listQuery{},
	}
	return ret
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
	q.isDebug = true
	return q
}

func (q *Query[T]) OrderDesc(order string) *Query[T] {
	q.listQuery.addOrder(order, true)
	return q
}

func (q *Query[T]) List() []*T {
	var items interface{}
	err := q.resource.listItems(&items, q.listQuery, q.isDebug)
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
	return countItems(q.resource.app.db, q.resource.getID(), q.listQuery, q.isDebug)
}
