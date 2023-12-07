package pragelastic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type BulkUpdater[T any] struct {
	index *Index[T]

	indexer esutil.BulkIndexer
}

func (index *Index[T]) UpdateSingle(item *T) error {
	id := getID(item)

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	//index.client.esclientNew.Update.

	_, err = index.client.esclientNew.Index(index.indexName(), strings.NewReader(string(data)), func(request *esapi.IndexRequest) {
		request.DocumentID = id
	})
	return err
}

func (index *Index[T]) UpdateBulk() (*BulkUpdater[T], error) {

	/*
		OnError can caouse possible memory leaks
		https://github.com/elastic/go-elasticsearch/issues/232
	*/

	indexer, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		NumWorkers:    1,
		FlushBytes:    1000000,
		FlushInterval: 5 * time.Second,
		Index:         index.indexName(),
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

	updater.indexer.Stats()

	/*
		OnFailure can cause possible memory leaks
		https://github.com/elastic/go-elasticsearch/issues/232
	*/

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
