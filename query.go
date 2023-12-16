package prago

import "context"

type QueryData[T any] struct {
	listQuery *listQuery
}

func Query[T any](app *App) *QueryData[T] {
	res := GetResource[T](app)
	ret := &QueryData[T]{
		listQuery: res.query(context.Background()),
	}
	return ret
}

func (q *QueryData[T]) Context(ctx context.Context) *QueryData[T] {
	q.listQuery.context = ctx
	return q
}

func (q *QueryData[T]) ID(id any) *T {
	ret := q.listQuery.ID(id)
	if ret == nil {
		return nil
	}
	return ret.(*T)
}

func (q *QueryData[T]) Is(name string, value interface{}) *QueryData[T] {
	q.listQuery.Is(name, value)
	return q
}

func (q *QueryData[T]) Where(condition string, values ...interface{}) *QueryData[T] {
	q.listQuery.where(condition, values...)
	return q
}

func (q *QueryData[T]) Limit(limit int64) *QueryData[T] {
	q.listQuery.Limit(limit)
	return q
}

func (q *QueryData[T]) Offset(offset int64) *QueryData[T] {
	q.listQuery.Offset(offset)
	return q
}

func (q *QueryData[T]) Debug() *QueryData[T] {
	q.listQuery.isDebug = true
	return q
}

func (q *QueryData[T]) Order(order string) *QueryData[T] {
	q.listQuery.addOrder(order, false)
	return q
}

func (q *QueryData[T]) OrderDesc(order string) *QueryData[T] {
	q.listQuery.addOrder(order, true)
	return q
}

func (q *QueryData[T]) List() []*T {
	items, err := q.listQuery.list()
	if err != nil {
		panic(err)
	}
	return items.([]*T)
}

func (q *QueryData[T]) First() *T {
	ret, ok := q.listQuery.First().(*T)
	if !ok {
		return nil
	}
	return ret
}

func (q *QueryData[T]) Count() (int64, error) {
	return q.listQuery.count()
}
