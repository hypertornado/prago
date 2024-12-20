package pragosearch

import (
	"strings"
	"testing"
)

func TestPragosearch(t *testing.T) {

	index := NewMemoryIndex()
	if err := index.SetAnalyzer("name", "czech"); err != nil {
		panic(err)
	}

	err := index.Index("1").Set("name", "jeníček a mařenka").Do()
	if err != nil {
		t.Fatal(err)
	}

	err = index.Index("2").Set("name", "červená karkulka").Do()
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
	if err := index.SetAnalyzer("name", "czech"); err != nil {
		panic(err)
	}

	index.Index("1").Set("name", "bar").Do()
	index.Index("2").Set("name", "foo").Do()
	index.Index("3").Set("name", "foo").Priority(10).Do()
	index.Index("4").Set("name", "foo").Priority(5).Do()
	index.Index("5").Set("name", "baz").Do()

	result := index.Query("foo").Do()
	if strings.Join(result.GetIDs(), ";") != "3;4;2" {
		t.Fatal(result.GetIDs())
	}

}
