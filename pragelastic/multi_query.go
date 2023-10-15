package pragelastic

import (
	"context"
	"fmt"
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

// TODO: fix it for real multiquery search
func (mq *MultiQuery[T]) Search() (ret []*ESSearchResult, err error) {

	for k, v := range mq.queries {
		res, err := v.SearchResult()
		if err != nil {
			return nil, fmt.Errorf("error while getting resuld %d: %s", k, err)
		}
		ret = append(ret, res)
	}
	return ret, nil
}
