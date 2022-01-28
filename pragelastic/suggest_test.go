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

func TestSuggestWithTags(t *testing.T) {
	index := prepareTestIndex[TestStructSuggestTags]()
	err := index.Update(&TestStructSuggestTags{
		ID: "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.Update(&TestStructSuggestTags{
		ID: "2",
		SuggestionField: Suggest{
			Input:  "Město moře stavení",
			Weight: 10,
			Contexts: map[string][]string{
				"Tags": {"A", "B"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = index.Update(&TestStructSuggestTags{
		ID: "3",
		SuggestionField: Suggest{
			Input:  "Město moře stavení",
			Weight: 10,
			Contexts: map[string][]string{
				"Tags": {""},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = index.Update(&TestStructSuggestTags{
		ID: "4",
		SuggestionField: Suggest{
			Input:  "Město moře stavení",
			Weight: 100,
			Contexts: map[string][]string{
				"Tags": {"B", "C"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().mustSuggest("mesto", map[string][]string{
		"Tags": {"A"},
	}))
	if res != "2" {
		t.Fatal(res)
	}

	res = getIDS(index.Query().mustSuggest("mesto", map[string][]string{
		"Tags": {"C"},
	}))
	if res != "4" {
		t.Fatal(res)
	}

	res = getIDS(index.Query().mustSuggest("mesto", map[string][]string{
		"Tags": {"B"},
	}))
	if res != "4,2" {
		t.Fatal(res)
	}
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

	res := getIDS(index.Query().mustSuggest("pan", nil))
	if res != "4,3" {
		t.Fatal(res)
	}
}
