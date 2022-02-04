package pragelastic

import "github.com/olivere/elastic/v7"

func (q *Query[T]) Aggregation(name string, aggregation elastic.Aggregation) *Query[T] {
	q.aggregations[name] = aggregation
	return q
}
