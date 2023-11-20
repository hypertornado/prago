package pragelastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	//esapi "github.com/elastic/go-elasticsearch/esapi"
)

type Index[T any] struct {
	client    *Client
	fields    []*field
	fieldsMap map[string]*field
}

func NewIndex[T any](client *Client) *Index[T] {

	if client == nil {
		return nil
	}
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
	_, err := index.client.esclientNew.Indices.Flush()
	return err
}

func (index *Index[T]) Refresh() error {
	_, err := index.client.esclientNew.Indices.Refresh()
	return err
}

func (index *Index[T]) Delete() error {
	_, err := index.client.esclientNew.Indices.Delete([]string{index.indexName()})
	return err
}

func (index *Index[T]) Get(id string) (*T, error) {
	response, err := index.client.esclientNew.GetSource(index.indexName(), id)
	if err != nil {
		return nil, err
	}

	var item T
	err = json.NewDecoder(response.Body).Decode(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (index *Index[T]) Count() (int64, error) {

	stats, err := index.client.GetStats()
	if err != nil {
		return -1, err
	}
	indice := stats.Indices[index.indexName()]
	if indice == nil {
		return -1, errors.New("cant find indice")
	}
	return indice.Total.Docs.Count, nil

}

func (index *Index[T]) DeleteItem(id string) error {
	_, err := index.client.esclientNew.Delete(index.indexName(), id)
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

	_, err := index.client.esclientNew.Indices.Create(index.indexName(), func(request *esapi.IndicesCreateRequest) {
		request.Body = strings.NewReader(str)
	})

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
			//"asciifolding",
			//"lowercase",
			"lowercase",
			"asciifolding",
			"czech_stop",
			"czech_keywords",
			"czech_stemmer",
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
