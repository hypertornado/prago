package prago

type Query2[T any] struct {
	resource *Resource[T]
	query    Query
}

func (resource *Resource[T]) Query() *Query2[T] {
	ret := &Query2[T]{
		resource: resource,
		query:    resource.Resource.app.Query(),
	}
	return ret
}

func (q *Query2[T]) Is(name string, value interface{}) *Query2[T] {
	newQ := q.query.Is(name, value)
	q.query = newQ
	return q
}

func (q *Query2[T]) Where(condition string, values ...interface{}) *Query2[T] {
	newQ := q.query.Where(condition, values...)
	q.query = newQ
	return q
}

func (q *Query2[T]) Limit(limit int64) *Query2[T] {
	newQ := q.query.Limit(limit)
	q.query = newQ
	return q
}

func (q *Query2[T]) Offset(limit int64) *Query2[T] {
	newQ := q.query.Offset(limit)
	q.query = newQ
	return q
}

func (q *Query2[T]) List() []*T {
	var items interface{}
	q.resource.Resource.newArrayOfItems(&items)
	err := q.query.Get(items)
	if err != nil {
		panic(err)
	}
	transformed, ok := items.(*[]*T)
	if !ok {
		panic("unexpected type")
	}
	return *transformed
}

func (q *Query2[T]) First() *T {
	items := q.List()
	if len(items) > 0 {
		return items[0]
	}
	return nil
}

func (q *Query2[T]) Count() (int64, error) {
	var item T
	return q.query.Count(&item)
}
