package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/olivere/elastic/v7"
)

type Query[T any] struct {
	index         *Index[T]
	sortField     string
	sortAsc       bool
	limit         int64
	offset        int64
	filterQueries []elastic.Query
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
		q.filterQueries = append(q.filterQueries, bq)
		return q
	}
	if f.Type == "text" {
		mq := elastic.NewMatchQuery(field, value)
		q.filterQueries = append(q.filterQueries, mq)
	} else {
		q.filterQueries = append(q.filterQueries, elastic.NewTermsQuery(field, value))
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

func (query *Query[T]) getSearchService() (*elastic.SearchService, error) {
	q := query.index.client.eclient.
		Search().
		Index(query.index.indexName())

	if query.sortField != "" {
		q = q.Sort(query.sortField, query.sortAsc)
	}

	bq := elastic.NewBoolQuery()
	bq.Filter(query.filterQueries...)

	ret := q.
		From(int(query.offset)).
		Size(int(query.limit)).
		Query(bq)
	return ret, nil
}

func (query *Query[T]) searchResult() (*elastic.SearchResult, error) {
	service, err := query.getSearchService()
	if err != nil {
		return nil, err
	}
	return service.Do(context.Background())
}

func (query *Query[T]) List() ([]*T, int64, error) {
	res, err := query.searchResult()
	if err != nil {
		return nil, -1, err
	}

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

func (query *Query[T]) mustList() []*T {
	list, _, err := query.List()
	if err != nil {
		panic(err)
	}
	return list
}
