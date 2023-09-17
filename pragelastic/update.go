package pragelastic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/olivere/elastic/v7"
)

type BulkUpdater[T any] struct {
	index     *Index[T]
	processor *elastic.BulkProcessor

	indexer esutil.BulkIndexer
}

func (index *Index[T]) UpdateSingleNew(item *T) error {
	bulk, err := index.UpdateBulk()
	if err != nil {
		return err
	}
	bulk.AddItem(item)
	return bulk.Close()
}

func (index *Index[T]) UpdateSingle(item *T) error {
	id := getID(item)

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_, err = index.client.esclientNew.Index(index.indexName(), strings.NewReader(string(data)), func(request *esapi.IndexRequest) {
		request.DocumentID = id
	})
	return err
}

func (index *Index[T]) UpdateBulk() (*BulkUpdater[T], error) {

	indexer, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      index.indexName(),
		ErrorTrace: true,
		OnError: func(ctx context.Context, err error) {
			panic(err)
		},
	})

	ret := &BulkUpdater[T]{
		index:   index,
		indexer: indexer,
	}

	return ret, nil
}

func (updater *BulkUpdater[T]) AddItem(item *T) {
	id := getID(item)

	data, err := json.Marshal(item)
	if err != nil {
		panic(err)
	}

	err = updater.indexer.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			DocumentID: id,
			Action:     "index",
			Body:       strings.NewReader(string(data)),
		},
	)
	if err != nil {
		panic(err)
	}
}

func (updater *BulkUpdater[T]) Close() error {
	return updater.indexer.Close(context.Background())
}
