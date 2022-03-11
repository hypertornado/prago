package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/olivere/elastic/v7"
)

type Query[T any] struct {
	index        *Index[T]
	sortField    string
	sortAsc      bool
	limit        int64
	offset       int64
	boolQuery    *elastic.BoolQuery
	context      context.Context
	aggregations map[string]elastic.Aggregation
}

func (index *Index[T]) Query() *Query[T] {
	q := &Query[T]{
		index:        index,
		boolQuery:    elastic.NewBoolQuery(),
		limit:        10,
		context:      context.Background(),
		aggregations: make(map[string]elastic.Aggregation),
	}
	return q
}

func (q *Query[T]) Context(ctx context.Context) *Query[T] {
	q.context = ctx
	return q
}

func (q *Query[T]) Filter(field string, value interface{}) *Query[T] {
	return q.FilterQuery(
		q.toQuery(field, value),
	)
}

func (q *Query[T]) Must(field string, value interface{}) *Query[T] {
	return q.MustQuery(
		q.toQuery(field, value),
	)
}

func (q *Query[T]) MustNot(field string, value interface{}) *Query[T] {
	return q.MustNotQuery(
		q.toQuery(field, value),
	)
}

func (q *Query[T]) Should(field string, value interface{}) *Query[T] {
	return q.ShouldQuery(
		q.toQuery(field, value),
	)
}

func (q *Query[T]) FilterQuery(query elastic.Query) *Query[T] {
	q.boolQuery.Filter(query)
	return q
}

func (q *Query[T]) MustQuery(query elastic.Query) *Query[T] {
	q.boolQuery.Must(query)
	return q
}

func (q *Query[T]) MustNotQuery(query elastic.Query) *Query[T] {
	q.boolQuery.MustNot(query)
	return q
}

func (q *Query[T]) ShouldQuery(query elastic.Query) *Query[T] {
	q.boolQuery.Should(query)
	return q
}

func (q *Query[T]) toQuery(field string, value interface{}) elastic.Query {
	f := q.index.fieldsMap[field]
	if f.Type == "keyword" && reflect.TypeOf(value) == reflect.TypeOf([]string{}) {
		bq := elastic.NewBoolQuery()
		shouldQueries := []elastic.Query{}
		arr := value.([]string)
		for _, v := range arr {
			shouldQueries = append(shouldQueries, elastic.NewTermQuery(field, v))
		}
		bq.Should(shouldQueries...)
		return bq
	}
	if f.Type == "text" {
		return elastic.NewMatchQuery(field, value)
	} else {
		return elastic.NewTermsQuery(field, value)
	}
}

func (q *Query[T]) Sort(fieldName string, asc bool) *Query[T] {
	q.sortField = fieldName
	q.sortAsc = asc
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

func (query *Query[T]) createSearchSource() *elastic.SearchSource {
	source := elastic.NewSearchSource()

	if query.sortField != "" {
		source = source.Sort(query.sortField, query.sortAsc)
	}

	for k, v := range query.aggregations {
		source.Aggregation(k, v)
	}

	source.
		From(int(query.offset)).
		Size(int(query.limit)).
		Query(query.boolQuery)

	return source
}

func (query *Query[T]) getSearchService() (*elastic.SearchService, error) {
	q := query.
		index.
		client.
		eclient.
		Search().
		Index(query.index.indexName())

	q.SearchSource(
		query.createSearchSource(),
	)
	return q, nil
}

func (query *Query[T]) Delete() error {
	deleteService := elastic.NewDeleteByQueryService(query.index.client.eclient)
	deleteService.Index(query.index.indexName())
	deleteService.Query(query.boolQuery)
	_, err := deleteService.Do(query.context)
	return err
}

func (query *Query[T]) SearchResult() (*elastic.SearchResult, error) {
	service, err := query.getSearchService()
	if err != nil {
		return nil, err
	}
	return service.Do(context.Background())
}

func (index Index[T]) SearchResultToList(res *elastic.SearchResult) ([]*T, int64, error) {
	var ret []*T
	for _, v := range res.Hits.Hits {
		var t T
		err := json.Unmarshal(v.Source, &t)
		if err != nil {
			return nil, -1, fmt.Errorf("can't unmarshal search result: %s", err)
		}
		ret = append(ret, &t)
	}
	return ret, res.Hits.TotalHits.Value, nil
}

func (query *Query[T]) List() ([]*T, int64, error) {
	res, err := query.SearchResult()
	if err != nil {
		return nil, -1, err
	}
	return query.index.SearchResultToList(res)
}

func (query *Query[T]) mustList() []*T {
	list, _, err := query.List()
	if err != nil {
		panic(err)
	}
	return list
}
