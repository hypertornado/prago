package pragelastic

import (
	"testing"
)

type TestStructSuggestTags struct {
	ID              string
	SuggestionField Suggest `elastic-analyzer:"czech_suggest" elastic-category-context-name:"Tags"`
}

type TestStructSuggestNoTags struct {
	ID              string
	SuggestionField Suggest `elastic-analyzer:"czech_suggest"`
}

func TestSuggestNoTags(t *testing.T) {
	index := prepareTestIndex[TestStructSuggestNoTags]()
	err := index.Update(&TestStructSuggestNoTags{
		ID: "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.Update(&TestStructSuggestNoTags{
		ID: "2",
		SuggestionField: Suggest{
			Input:  "Město moře stavení",
			Weight: 10,
			/*Contexts: map[string][]string{
				"Tags": {"hello world dd"},
			},*/
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	index.Update(&TestStructSuggestNoTags{
		ID: "3",
		SuggestionField: Suggest{
			Input:  "Pán hrad stavení",
			Weight: 100,
		},
	})
	index.Update(&TestStructSuggestNoTags{
		ID: "4",
		SuggestionField: Suggest{
			Input:  "pan hrad stavení",
			Weight: 1000,
		},
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().mustSuggest("pan"))
	if res != "4,3" {
		t.Fatal(res)
	}
}
