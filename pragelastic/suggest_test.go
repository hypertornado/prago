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
	err := index.UpdateSingle(&TestStructSuggestTags{
		ID: "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.UpdateSingle(&TestStructSuggestTags{
		ID: "2",
		SuggestionField: Suggest{
			Input:  []string{"Město moře stavení"},
			Weight: 10,
			Contexts: map[string][]string{
				"Tags": {"A", "B"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = index.UpdateSingle(&TestStructSuggestTags{
		ID: "3",
		SuggestionField: Suggest{
			Input:  []string{"Město moře stavení"},
			Weight: 10,
			Contexts: map[string][]string{
				"Tags": {""},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = index.UpdateSingle(&TestStructSuggestTags{
		ID: "4",
		SuggestionField: Suggest{
			Input:  []string{"Město moře stavení"},
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

	res := getIDS(index.mustSuggest("mesto", map[string][]string{
		"Tags": {"A"},
	}))
	if res != "2" {
		t.Fatal(res)
	}

	res = getIDS(index.mustSuggest("mesto", map[string][]string{
		"Tags": {"C"},
	}))
	if res != "4" {
		t.Fatal(res)
	}

	res = getIDS(index.mustSuggest("mesto", map[string][]string{
		"Tags": {"B"},
	}))
	if res != "4,2" {
		t.Fatal(res)
	}
}

func TestSuggestNoTags(t *testing.T) {
	index := prepareTestIndex[TestStructSuggestNoTags]()
	err := index.UpdateSingle(&TestStructSuggestNoTags{
		ID: "1",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.UpdateSingle(&TestStructSuggestNoTags{
		ID: "2",
		SuggestionField: Suggest{
			Input:  []string{"Město", "moře", "stavení"},
			Weight: 10,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	index.UpdateSingle(&TestStructSuggestNoTags{
		ID: "3",
		SuggestionField: Suggest{
			Input:  []string{"Pán", "hrad", "stavení", "brno"},
			Weight: 100,
		},
	})
	index.UpdateSingle(&TestStructSuggestNoTags{
		ID: "4",
		SuggestionField: Suggest{
			Input:  []string{"pan", "hrad", "stavení"},
			Weight: 1000,
		},
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.mustSuggest("pan", nil))
	if res != "4,3" {
		t.Fatal(res)
	}

	res = getIDS(index.mustSuggest("brn", nil))
	if res != "3" {
		t.Fatal(res)
	}
}
