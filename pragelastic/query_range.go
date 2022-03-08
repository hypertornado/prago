package pragelastic

import "github.com/olivere/elastic/v7"

func (q *Query[T]) Range(name string, from, to any) *Query[T] {
	rangeQuery := elastic.NewRangeQuery(name).Gte(from).Lte(to)
	q.FilterQuery(rangeQuery)
	return q
}

func (q *Query[T]) LowerThanOrEqual(name string, value any) *Query[T] {
	rangeQuery := elastic.NewRangeQuery(name).Lte(value)
	q.FilterQuery(rangeQuery)
	return q
}

func (q *Query[T]) GreaterThanOrEqual(name string, value any) *Query[T] {
	rangeQuery := elastic.NewRangeQuery(name).Gte(value)
	q.FilterQuery(rangeQuery)
	return q
}
