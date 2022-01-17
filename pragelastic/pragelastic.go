package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/olivere/elastic/v7"
)

type Client struct {
	prefix  string
	eclient *elastic.Client
}

func New(id string) *Client {
	client, err := elastic.NewClient()
	if err != nil {
		panic(err)
	}
	return &Client{
		prefix:  id,
		eclient: client,
	}
}

func NewIndex[T any](client *Client) *Index[T] {
	ret := &Index[T]{
		client: client,
	}
	return ret
}

type Index[T any] struct {
	//suffix string
	client *Client
}

func (index Index[T]) Flush() error {
	_, err := index.client.eclient.Flush().Do(context.Background())
	return err
}

func (index Index[T]) Delete() error {
	_, err := index.client.eclient.DeleteIndex(index.indexName()).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (index Index[T]) Get(id int64) (*T, error) {
	idStr := fmt.Sprintf("%d", id)
	res, err := index.client.eclient.Get().Index(index.indexName()).Id(idStr).Do(context.Background())
	if err != nil {
		return nil, err
	}

	var item T
	err = json.Unmarshal(res.Source, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (index Index[T]) count() (int64, error) {
	stats, err := index.client.eclient.IndexStats(index.indexName()).Do(context.Background())
	if err != nil {
		return -1, err
	}
	return stats.All.Primaries.Docs.Count, nil
}

//func (index Index[T]) Search(item *T) error {

//}

func (index Index[T]) DeleteItem(id int64) error {
	idStr := fmt.Sprintf("%d", id)
	_, err := index.client.eclient.Delete().Index(index.indexName()).Id(idStr).Do(context.Background())
	return err
}

func (index Index[T]) Update(item *T) error {
	id := getID(item)
	idStr := fmt.Sprintf("%d", id)
	_, err := index.client.eclient.Index().Index(index.indexName()).BodyJson(item).Id(idStr).Do(context.Background())
	return err
}

func (index Index[T]) Create() error {
	str := index.indexDataStr()
	_, err := index.client.eclient.CreateIndex(index.indexName()).BodyString(str).Do(context.Background())
	return err
}

func (index Index[T]) indexName() string {
	var item T
	t := reflect.TypeOf(item)
	suffix := t.Name()
	suffix = strings.ToLower(suffix)
	return fmt.Sprintf("%s-%s", index.client.prefix, suffix)
}

func (index Index[T]) indexData() interface{} {
	properties := map[string]any{}

	for _, v := range getFields[T]() {
		properties[v.Name] = map[string]any{
			"type": v.Type,
		}
	}

	ret := map[string]any{
		"mappings": map[string]any{
			"properties": properties,
		},
	}
	return ret
}

func (index Index[T]) indexDataStr() string {
	data, err := json.MarshalIndent(index.indexData(), "", " ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

type field struct {
	Name string
	Type string
}

//func (index Index[T]) Add(item *T) string {

//}
