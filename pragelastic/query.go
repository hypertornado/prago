package pragelastic

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/olivere/elastic/v7"
)

type Query[T any] struct {
	index     *Index[T]
	sortField string
	sortAsc   bool
	limit     int64
	offset    int64
	queries   []elastic.Query
}

func (index *Index[T]) Query() *Query[T] {
	q := &Query[T]{
		index: index,
		limit: 10,
	}
	return q
}

func (q *Query[T]) Is(field string, value interface{}) *Query[T] {
	f := q.index.fieldsMap[field]
	if f.Type == "keyword" && reflect.TypeOf(value) == reflect.TypeOf([]string{}) {
		bq := elastic.NewBoolQuery()
		shouldQueries := []elastic.Query{}
		arr := value.([]string)
		for _, v := range arr {
			shouldQueries = append(shouldQueries, elastic.NewTermQuery(field, v))
		}
		bq.Should(shouldQueries...)
		q.queries = append(q.queries, bq)
		return q
	}
	if f.Type == "text" {
		mq := elastic.NewMatchQuery(field, value)
		q.queries = append(q.queries, mq)
	} else {
		q.queries = append(q.queries, elastic.NewTermsQuery(field, value))
	}
	return q
}

func (q *Query[T]) Sort(fieldName string, desc bool) *Query[T] {
	q.sortField = fieldName
	q.sortAsc = desc
	return q
}

func (q *Query[T]) Limit(limit int64) *Query[T] {
	q.limit = limit
	return q
}

func (q *Query[T]) Offset(offset int64) *Query[T] {
	q.offset = offset
	return q
}

func (query *Query[T]) List() []*T {
	q := query.index.client.eclient.
		Search().
		Index(query.index.indexName())

	if query.sortField != "" {
		q = q.Sort(query.sortField, query.sortAsc)
	}

	bq := elastic.NewBoolQuery()
	for _, v := range query.queries {
		bq.Must(
			v,
		)
	}

	res, err := q.
		From(int(query.offset)).
		Size(int(query.limit)).
		Query(bq).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	var ret []*T

	for _, v := range res.Hits.Hits {
		var t T
		err := json.Unmarshal(v.Source, &t)
		if err != nil {
			panic(err)
		}
		ret = append(ret, &t)
	}
	return ret
}
