package pragosearch

import (
	"strings"
	"testing"
)

func TestPragosearch(t *testing.T) {

	index := NewMemoryIndex()
	index.Field("name").Analyzer("czech")

	err := index.Add("1").Set("name", "jeníček a mařenka").Do()
	if err != nil {
		t.Fatal(err)
	}

	err = index.Add("2").Set("name", "červená karkulka").Do()
	if err != nil {
		t.Fatal(err)
	}

	result := index.Query("karkulka").Do()
	if result.Total != 1 {
		t.Fatal(result.Total)
	}

	if strings.Join(result.GetIDs(), ";") != "2" {
		t.Fatal(result.GetIDs())
	}
}

func TestDocPriority(t *testing.T) {
	index := NewMemoryIndex()
	index.Field("name").Analyzer("czech")

	index.Add("1").Set("name", "bar").Do()
	index.Add("2").Set("name", "foo").Do()
	index.Add("3").Set("name", "foo").Priority(10).Do()
	index.Add("4").Set("name", "foo").Priority(5).Do()
	index.Add("5").Set("name", "baz").Do()

	result := index.Query("foo").Do()
	if strings.Join(result.GetIDs(), ";") != "3;4;2" {
		t.Fatal(result.GetIDs())
	}
}

func TestFieldPriority(t *testing.T) {
	index := NewMemoryIndex()
	index.Field("name").Priority(100)
	index.Field("description").Priority(1)

	index.Add("1").Set("name", "foo1").Do()
	index.Add("2").Set("name", "foo2 bar bar").Do()
	index.Add("3").Set("name", "foo3").Set("description", "bar").Do()
	index.Add("4").Set("name", "foo4 bar").Do()

	result := index.Query("bar").Do()
	if strings.Join(result.GetIDs(), ";") != "2;4;3" {
		t.Fatal(result.GetIDs())
	}
}

func TestDataStore(t *testing.T) {
	index := NewMemoryIndex()
	index.Field("name")

	index.Add("1").Set("name", "foo").StoreData("baz").Do()

	result := index.Query("foo").Do()
	if result.Results[0].Data != "baz" {
		t.Fatal(result.Results[0].Data)
	}

}

func TestSuggestion(t *testing.T) {
	index := NewMemoryIndex()
	index.Field("name")

	index.Add("1").Set("name", "foo").Do()
	index.Add("2").Set("name", "barbar").Do()
	index.Add("3").Set("name", "baz").Do()

	result := index.Suggest("bar").Do()
	if strings.Join(result.GetIDs(), ";") != "2" {
		t.Fatal(result.GetIDs())
	}
}

func TestSuggestionCH(t *testing.T) {

	index := NewMemoryIndex()
	index.Field("name")

	index.Add("1").Set("name", "foo").Do()
	index.Add("2").Set("name", "chariclea").Do()
	index.Add("3").Set("name", "baz").Do()

	result := index.Suggest("c").Do()
	if strings.Join(result.GetIDs(), ";") != "2" {
		t.Fatal(result.GetIDs())
	}
}

func TestSuggestionEmpty(t *testing.T) {
	index := NewMemoryIndex()
	index.Field("name")

	index.Add("1").Set("name", "foo").Do()
	index.Add("2").Set("name", "bar").Priority(100).Do()
	index.Add("3").Set("name", "baz").Priority(10).Do()

	result := index.Suggest("").Do()
	if strings.Join(result.GetIDs(), ";") != "2;3;1" {
		t.Fatal(result.GetIDs())
	}
}
