package pragelastic

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type MultiQuery[T any] struct {
	index   *Index[T]
	context context.Context
	queries []*Query[T]
}

func (index *Index[T]) MultiQuery() *MultiQuery[T] {
	return &MultiQuery[T]{
		index:   index,
		context: context.Background(),
	}
}

func (mq *MultiQuery[T]) Add(q ...*Query[T]) *MultiQuery[T] {
	mq.queries = append(mq.queries, q...)
	return mq
}

func (mq *MultiQuery[T]) Context(ctx context.Context) *MultiQuery[T] {
	mq.context = ctx
	return mq
}

func (mq *MultiQuery[T]) Search() ([]*elastic.SearchResult, error) {
	multi := elastic.NewMultiSearchService(mq.index.client.esclientOld)
	for _, v := range mq.queries {
		request := elastic.NewSearchRequest()
		request.Index(v.index.indexName())
		request.SearchSource(v.createSearchSource())
		multi.Add(request)
	}

	res, err := multi.Do(mq.context)
	if err != nil {
		return nil, err
	}
	return res.Responses, nil
}
