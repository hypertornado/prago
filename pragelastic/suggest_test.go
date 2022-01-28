package pragelastic

import "testing"

type TestStructSuggest struct {
	ID              string
	SuggestionField Suggest `elastic-analyzer:"czech_suggest" elastic-category-context-name:"Tags"`
}

func TestSuggest(t *testing.T) {
	if true {
		return
	}

	index := prepareTestIndex[TestStructSuggest]()
	index.Update(&TestStructSuggest{
		ID: "1",
	})

	index.Update(&TestStructSuggest{
		ID: "2",
		SuggestionField: Suggest{
			Input:  "Město moře stavení",
			Weight: 10,
		},
	})
	index.Update(&TestStructSuggest{
		ID: "3",
		SuggestionField: Suggest{
			Input:  "Pán hrad stavení",
			Weight: 100,
		},
	})
	index.Update(&TestStructSuggest{
		ID: "4",
		SuggestionField: Suggest{
			Input:  "pan hrad stavení",
			Weight: 1000,
		},
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().mustSuggest("mesto"))
	if res != "2" {
		t.Fatal(res)
	}

	/*
		res = getIDS(index.Query().mustSuggest("pan"))
		if res != "4,3" {
			t.Fatal(res)
		}*/

	/*res = getIDS(index.Query().Is("SomeCount", 200).mustSuggest("pan"))
	if res != "4" {
		t.Fatal(res)
	}*/

}
