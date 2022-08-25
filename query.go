package prago

type Query[T any] struct {
	listQuery *listQuery
}

func (resource *Resource[T]) Query() *Query[T] {
	ret := &Query[T]{
		listQuery: resource.data.query(),
	}
	return ret
}

func (resource *Resource[T]) ID(id any) *T {
	return resource.Query().ID(id)
}

func (q *Query[T]) ID(id any) *T {
	return q.listQuery.ID(id).(*T)
}

func (q *Query[T]) Is(name string, value interface{}) *Query[T] {
	q.listQuery.Is(name, value)
	return q
}

func (q *Query[T]) Where(condition string, values ...interface{}) *Query[T] {
	q.listQuery.where(condition, values...)
	return q
}

func (q *Query[T]) Limit(limit int64) *Query[T] {
	q.listQuery.Limit(limit)
	return q
}

func (q *Query[T]) Offset(offset int64) *Query[T] {
	q.listQuery.Offset(offset)
	return q
}

func (q *Query[T]) Debug() *Query[T] {
	q.listQuery.isDebug = true
	return q
}

func (q *Query[T]) Order(order string) *Query[T] {

	q.listQuery.addOrder(order, false)
	return q
}

func (q *Query[T]) OrderDesc(order string) *Query[T] {
	q.listQuery.addOrder(order, true)
	return q
}

func (q *Query[T]) List() []*T {
	items, err := q.listQuery.list()
	if err != nil {
		panic(err)
	}
	return items.([]*T)
}

func (q *Query[T]) First() *T {
	ret, ok := q.listQuery.First().(*T)
	if !ok {
		return nil
	}
	return ret

	/*items := q.List()
	if len(items) > 0 {
		return items[0]
	}
	return nil*/
}

func (q *Query[T]) Count() (int64, error) {
	return q.listQuery.count()
}
