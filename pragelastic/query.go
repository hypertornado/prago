package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

//https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/examples.html#search

type Query[T any] struct {
	index        *Index[T]
	sortField    string
	sortAsc      bool
	limit        int64
	offset       int64
	boolQuery    *ESBoolQuery
	context      context.Context
	aggregations map[string]ESAggregation
}

func (index *Index[T]) Query() *Query[T] {

	q := &Query[T]{
		index:        index,
		boolQuery:    NewESBoolQuery(),
		limit:        10,
		context:      context.Background(),
		aggregations: make(map[string]ESAggregation),
	}
	return q
}

// TODO: make context mandatory
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

func (q *Query[T]) FilterQuery(query ESQuery) *Query[T] {
	q.boolQuery.Filter(query)
	return q
}

func (q *Query[T]) MustQuery(query ESQuery) *Query[T] {
	q.boolQuery.Must(query)
	return q
}

func (q *Query[T]) MustNotQuery(query ESQuery) *Query[T] {
	q.boolQuery.MustNot(query)
	return q
}

func (q *Query[T]) ShouldQuery(query ESQuery) *Query[T] {
	q.boolQuery.Should(query)
	return q
}

func (q *Query[T]) toQuery(field string, value interface{}) ESQuery {
	fieldName, _, _ := strings.Cut(field, ".")
	f := q.index.fieldsMap[fieldName]
	if f == nil {
		panic("could not find field with name: " + fieldName)
	}
	if f.Type == "keyword" && reflect.TypeOf(value) == reflect.TypeOf([]string{}) {
		bq := NewESBoolQuery()
		shouldQueries := []ESQuery{}
		arr := value.([]string)
		for _, v := range arr {
			shouldQueries = append(shouldQueries, NewESTermQuery(field, v))
		}
		bq.Should(shouldQueries...)
		return bq
	}
	if f.Type == "text" {
		return NewESMatchQuery(field, value)
	} else {
		return NewESTermsQuery(field, value)
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

func (query *Query[T]) createSearchSource() *ESSearchSource {

	source := NewESSearchSource()

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

func (query *Query[T]) Delete() error {
	qSource, err := query.boolQuery.Source()
	if err != nil {
		return err
	}
	var srcData = map[string]any{
		"query": qSource,
	}

	data, err := json.Marshal(srcData)
	if err != nil {
		panic(err)
	}

	resp, err := query.index.client.esclientNew.DeleteByQuery([]string{query.index.indexName()}, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf(resp.String())
	}
	return nil

}

func (query *Query[T]) SearchResult() (*ESSearchResult, error) {

	ss := query.createSearchSource()
	srcData, err := ss.Source()
	if err != nil {
		return nil, err
	}

	res, err := query.index.client.esclientNew.Search(func(sr *esapi.SearchRequest) {

		sr.Index = []string{query.index.indexName()}

		data, err := json.Marshal(srcData)
		if err != nil {
			panic(err)
		}

		reader := strings.NewReader(string(data))
		sr.Body = reader
	})

	if err != nil {
		return nil, err
	}

	var result ESSearchResult

	//TODO: faster unmarshal
	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (index Index[T]) SearchResultToList(res *ESSearchResult) ([]*T, int64, error) {
	var ret []*T

	if res.Hits == nil {
		return ret, 0, nil
	}
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
