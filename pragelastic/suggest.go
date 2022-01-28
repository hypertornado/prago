package pragelastic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
)

type Suggest struct {
	Input  string `json:"input"`
	Weight int64  `json:"weight"`
}

type SuggestCategoryContext struct {
	Name string `json:"name"`
	Type string `json:"type"`
	//Path string `json:"path"`
}

func (index *Index[T]) getSuggestFieldName() string {
	for _, v := range index.fields {
		if v.Type == "completion" {
			return v.Name
		}
	}
	return ""
}

func (query *Query[T]) mustSuggest(q string) []*T {
	ret, err := query.Suggest(q)
	if err != nil {
		panic(err)
	}
	return ret
}

func (query *Query[T]) Suggest(q string) ([]*T, error) {
	fieldName := query.index.getSuggestFieldName()
	if fieldName == "" {
		return nil, fmt.Errorf("Can't find suggest field name: no field has type copletion")
	}

	suggesterName := "_suggester"
	cs := elastic.NewCompletionSuggester(suggesterName).
		Field(fieldName).
		Prefix(q).
		SkipDuplicates(true)

	//cs.ContextQueries(query.filterQueries...)

	ss, err := query.getSearchService()
	if err != nil {
		return nil, err
	}

	ss.Suggester(cs)

	searchResult, err := ss.Do(context.Background())
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
