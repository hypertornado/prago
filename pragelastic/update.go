package pragelastic

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type BulkUpdater[T any] struct {
	index     *Index[T]
	processor *elastic.BulkProcessor
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
	_, err := index.
		client.
		esclientOld.
		Index().
		Index(
			index.indexName(),
		).
		BodyJson(item).
		Id(id).
		Do(
			context.Background(),
		)
	return err
}

func (index *Index[T]) UpdateBulk() (*BulkUpdater[T], error) {
	bp, err := index.client.esclientOld.BulkProcessor().Name("prago-bulk-updater").
		Before(func(executionId int64, requests []elastic.BulkableRequest) {
		}).
		After(func(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
			/*for _, v := range response.Failed() {
				fmt.Println(v.Error)
			}*/
		}).
		Workers(2).Do(context.Background())
	if err != nil {
		return nil, err
	}
	err = bp.Start(context.Background())
	if err != nil {
		return nil, err
	}

	ret := &BulkUpdater[T]{
		index:     index,
		processor: bp,
	}
	return ret, nil
}

func (updater *BulkUpdater[T]) AddItem(item *T) {
	id := getID(item)
	r := elastic.NewBulkIndexRequest()
	r.
		Index(updater.index.indexName()).
		Id(id).
		Doc(item)
	updater.processor.Add(r)
}

func (updater *BulkUpdater[T]) Close() error {
	err := updater.processor.Flush()
	if err != nil {
		return err
	}
	return updater.processor.Close()
}
