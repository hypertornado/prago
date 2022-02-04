package pragelastic

import "github.com/olivere/elastic/v7"

func (q *Query[T]) Range(name string, from, to int64) *Query[T] {
	rangeQuery := elastic.NewRangeQuery(name).Gte(from).Lte(to)
	q.FilterQuery(rangeQuery)
	return q
}
