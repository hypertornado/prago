package pragelastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type Suggest struct {
	Input    []string            `json:"input"`
	Weight   int64               `json:"weight"`
	Contexts map[string][]string `json:"contexts,omitempty"`
}

type SuggestCategoryContext struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (index *Index[T]) getSuggestFieldName() string {
	for _, v := range index.fields {
		if v.Type == "completion" {
			return v.Name
		}
	}
	return ""
}

func (index *Index[T]) mustSuggest(q string, categoryContexts map[string][]string) []*T {
	ret, err := index.Suggest(q, 10, categoryContexts)
	if err != nil {
		panic(err)
	}
	return ret
}

func (index *Index[T]) Suggest(q string, size int64, categoryContexts map[string][]string) ([]*T, error) {
	fieldName := index.getSuggestFieldName()
	if fieldName == "" {
		return nil, errors.New("can't find suggest field name: no field has type completion")
	}

	suggesterName := "_suggester"
	completionSuggester := NewESCompletionSuggester(suggesterName).
		Field(fieldName).
		Prefix(q).Size(int(size))

	for k, v := range categoryContexts {
		completionSuggester.ContextQuery(NewESSuggesterCategoryQuery(k, v...))
	}

	srcData, err := completionSuggester.Source(true)
	if err != nil {
		return nil, err
	}

	res, err := index.client.esclientNew.Search(func(sr *esapi.SearchRequest) {
		sr.Index = []string{index.indexName()}

		var rootSrc = map[string]any{
			"suggest": srcData,
		}

		data, err := json.Marshal(rootSrc)
		if err != nil {
			panic(err)
		}

		reader := strings.NewReader(string(data))
		sr.Body = reader
	})

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("error while searching for suggest: %s", res.Status())
	}

	var result ESSearchResult
	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	var ret []*T
	for _, v := range result.Suggest[suggesterName] {
		for _, option := range v.Options {

			var t T
			err := json.Unmarshal(option.Source, &t)
			if err != nil {
				panic(err)
			}
			ret = append(ret, &t)
		}
	}
	return ret, nil
}
