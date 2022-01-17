package prago

type Query[T any] struct {
	resource *Resource[T]
	query    query
}

func (resource *Resource[T]) Query() *Query[T] {
	ret := &Query[T]{
		resource: resource,
		query:    resource.query(),
	}
	return ret
}

func (q *Query[T]) Is(name string, value interface{}) *Query[T] {
	newQ := q.query.is(name, value)
	q.query = newQ
	return q
}

func (q *Query[T]) Where(condition string, values ...interface{}) *Query[T] {
	newQ := q.query.where(condition, values...)
	q.query = newQ
	return q
}

func (q *Query[T]) Limit(limit int64) *Query[T] {
	newQ := q.query.limit(limit)
	q.query = newQ
	return q
}

func (q *Query[T]) Offset(limit int64) *Query[T] {
	newQ := q.query.offset(limit)
	q.query = newQ
	return q
}

func (q *Query[T]) Order(order string) *Query[T] {
	newQ := q.query.order(order)
	q.query = newQ
	return q
}

func (q *Query[T]) Debug() *Query[T] {
	newQ := q.query.debug()
	q.query = newQ
	return q
}

func (q *Query[T]) OrderDesc(order string) *Query[T] {
	newQ := q.query.orderDesc(order)
	q.query = newQ
	return q
}

func (q *Query[T]) List() []*T {
	items, err := q.query.list()
	if err != nil {
		panic(err)
	}
	transformed, ok := items.([]*T)
	if !ok {
		panic("unexpected type")
	}
	return transformed

}

func (q *Query[T]) First() *T {
	items := q.List()
	if len(items) > 0 {
		return items[0]
	}
	return nil
}

func (q *Query[T]) Count() (int64, error) {
	return q.query.count()
}
