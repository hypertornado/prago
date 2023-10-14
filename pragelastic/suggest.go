package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
)

type Suggest struct {
	Input    string              `json:"input"`
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
	ret, err := index.Suggest(q, categoryContexts)
	if err != nil {
		panic(err)
	}
	return ret
}

func (index *Index[T]) Suggest(q string, categoryContexts map[string][]string) ([]*T, error) {
	fieldName := index.getSuggestFieldName()
	if fieldName == "" {
		return nil, fmt.Errorf("Can't find suggest field name: no field has type completion")
	}

	suggesterName := "_suggester"
	completionSuggester := elastic.NewCompletionSuggester(suggesterName).
		Field(fieldName).
		Prefix(q)

	for k, v := range categoryContexts {
		completionSuggester.ContextQuery(elastic.NewSuggesterCategoryQuery(k, v...))
	}

	searchService := index.client.esclientOld.
		Search().
		Index(index.indexName()).
		Suggester(completionSuggester)

	searchResult, err := searchService.Do(context.Background())
	if err != nil {
		return nil, err
	}

	var ret []*T
	suggestions := searchResult.Suggest[suggesterName]
	for _, v := range suggestions {
		for _, option := range v.Options {
			var t T
			err := json.Unmarshal(option.Source, &t)
			if err != nil {
				return nil, fmt.Errorf("can't unmarshal suggestion result: %s", err)
			}
			ret = append(ret, &t)

		}
	}
	return ret, nil
}
