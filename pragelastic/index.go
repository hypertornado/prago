package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Index[T any] struct {
	client    *Client
	fields    []*field
	fieldsMap map[string]*field
}

func NewIndex[T any](client *Client) *Index[T] {
	ret := &Index[T]{
		client:    client,
		fields:    getFields[T](),
		fieldsMap: make(map[string]*field),
	}
	for _, v := range ret.fields {
		ret.fieldsMap[v.Name] = v
	}
	return ret
}

func (index *Index[T]) Flush() error {
	_, err := index.client.eclient.Flush().Do(context.Background())
	return err
}

func (index *Index[T]) Refresh() error {
	_, err := index.client.eclient.Refresh().Do(context.Background())
	return err
}

func (index *Index[T]) Delete() error {
	_, err := index.client.eclient.DeleteIndex(index.indexName()).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (index *Index[T]) Get(id string) (*T, error) {
	res, err := index.client.eclient.Get().Index(index.indexName()).Id(id).Do(context.Background())
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

func (index *Index[T]) Count() (int64, error) {
	stats, err := index.client.eclient.IndexStats(index.indexName()).Do(context.Background())
	if err != nil {
		return -1, err
	}
	return stats.All.Primaries.Docs.Count, nil
}

func (index *Index[T]) DeleteItem(id string) error {
	_, err := index.client.eclient.Delete().Index(index.indexName()).Id(id).Do(context.Background())
	return err
}

func (index *Index[T]) indexName() string {
	var item T
	t := reflect.TypeOf(item)
	suffix := t.Name()
	suffix = strings.ToLower(suffix)
	return fmt.Sprintf("%s-%s", index.client.prefix, suffix)
}

func (index *Index[T]) Create() error {
	str := index.indexDataStr()
	//fmt.Println("creating index:", index.indexName())
	_, err := index.client.eclient.CreateIndex(index.indexName()).BodyString(str).Do(context.Background())
	return err
}

func (index *Index[T]) getSettings() map[string]interface{} {
	settings := make(map[string]interface{})

	analysis := make(map[string]interface{})

	filter := make(map[string]interface{})
	filter["czech_stop"] = map[string]any{
		"type":      "stop",
		"stopwords": "_czech_",
	}
	filter["czech_keywords"] = map[string]any{
		"type":     "keyword_marker",
		"keywords": []string{"a"},
	}
	filter["czech_stemmer"] = map[string]any{
		"type":     "stemmer",
		"language": "czech",
	}

	analysis["filter"] = filter

	analyzer := make(map[string]interface{})
	analyzer["czech"] = map[string]any{
		"tokenizer": "standard",
		"filter": []string{
			"lowercase",
			"asciifolding",
			"czech_stop",
			"czech_keywords",
			"czech_stemmer",
		},
	}
	analyzer["czech_suggest"] = map[string]any{
		"tokenizer": "standard",
		"filter": []string{
			"lowercase",
			"asciifolding",
		},
	}
	analysis["analyzer"] = analyzer

	settings["analysis"] = analysis
	return settings
}

func (index Index[T]) indexData() interface{} {
	properties := map[string]any{}

	for _, v := range index.fields {
		property := map[string]any{
			"type": v.Type,
		}
		if v.Analyzer != "" {
			property["analyzer"] = v.Analyzer
		}
		if len(v.CategoryContexts) > 0 {
			property["contexts"] = v.CategoryContexts
		}
		if !v.Enabled {
			property["enabled"] = false
		}
		properties[v.Name] = property
	}

	ret := map[string]any{
		"settings": index.getSettings(),
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
