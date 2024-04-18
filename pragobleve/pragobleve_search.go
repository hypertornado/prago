package pragobleve

import (
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve/v2"
)

type PragoBleveQuery[T any] struct {
	pdIndex *PragoBleveIndex[T]

	phraseQueries []*phraseQuery
	boolQueries   []*boolQuery

	sortFields []string

	offset int64
	size   int64
}

type phraseQuery struct {
	FieldName string
	Q         string
}

type boolQuery struct {
	FieldName string
	Value     bool
}

func (index *PragoBleveIndex[T]) Query() *PragoBleveQuery[T] {
	return &PragoBleveQuery[T]{
		pdIndex: index,
		size:    -1,
	}
}

func (pq *PragoBleveQuery[T]) Phrase(field, q string) *PragoBleveQuery[T] {
	pq.phraseQueries = append(pq.phraseQueries, &phraseQuery{
		FieldName: field,
		Q:         q,
	})
	return pq
}

func (pq *PragoBleveQuery[T]) Bool(field string, value bool) *PragoBleveQuery[T] {
	pq.boolQueries = append(pq.boolQueries, &boolQuery{
		FieldName: field,
		Value:     value,
	})
	return pq
}

func (pq *PragoBleveQuery[T]) Sort(field string) *PragoBleveQuery[T] {
	pq.sortFields = append(pq.sortFields, field)
	return pq
}

func (pq *PragoBleveQuery[T]) SortDesc(field string) *PragoBleveQuery[T] {
	pq.sortFields = append(pq.sortFields, "-"+field)
	return pq
}

func (pq *PragoBleveQuery[T]) Offset(offset int64) *PragoBleveQuery[T] {
	pq.offset = offset
	return pq
}

func (pq *PragoBleveQuery[T]) Size(size int64) *PragoBleveQuery[T] {
	pq.size = size
	return pq
}

func (pq *PragoBleveQuery[T]) Search() ([]*T, error) {
	cq := bleve.NewConjunctionQuery()
	cq.AddQuery(bleve.NewMatchAllQuery())

	for _, phraseQuery := range pq.phraseQueries {
		cq.AddQuery(bleve.NewPhraseQuery([]string{phraseQuery.Q}, phraseQuery.FieldName))
	}
	for _, boolQuery := range pq.boolQueries {
		bfq := bleve.NewBoolFieldQuery(boolQuery.Value)
		bfq.SetField(boolQuery.FieldName)
		cq.AddQuery(bfq)
	}

	searchRequest := bleve.NewSearchRequest(cq)

	if pq.sortFields != nil {
		searchRequest.SortBy(pq.sortFields)
	}

	if pq.size >= 0 {
		searchRequest.Size = int(pq.size)
	}
	searchRequest.From = int(pq.offset)

	//searchRequest.AddFacet()
	fr := bleve.NewFacetRequest("XXx", 1000)
	fmt.Println(fr)
	//fr.a

	searchRequest.Fields = append(searchRequest.Fields, dataSaveField)
	searchResult, err := pq.pdIndex.bIndex.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var ret []*T

	for _, v := range searchResult.Hits {
		var item T
		must(json.Unmarshal([]byte(v.Fields[dataSaveField].(string)), &item))
		ret = append(ret, &item)
	}

	return ret, nil

}
