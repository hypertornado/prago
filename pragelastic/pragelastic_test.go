package pragelastic

import (
	"strings"
	"testing"
)

const testClientName = "pragelastic-test"

type TestStruct struct {
	ID        string
	Name      string `elastic-datatype:"keyword"`
	Text      string `elastic-datatype:"text" elastic-analyzer:"czech"`
	SomeCount int64
	IsOK      bool
	Tags      []string
}

func getIDS(items []*TestStruct) string {
	ret := []string{}
	for _, v := range items {
		ret = append(ret, v.ID)
	}
	return strings.Join(ret, ",")
}

func prepareTestIndex() *Index[TestStruct] {
	lib := New(testClientName)
	index := NewIndex[TestStruct](lib)
	index.Delete()
	err := index.Create()
	if err != nil {
		panic(err)
	}
	err = index.Flush()
	if err != nil {
		panic(err)
	}
	return index
}

func TestMultipleBooleanTags(t *testing.T) {
	index := prepareTestIndex()
	index.Update(&TestStruct{
		ID:        "1",
		Name:      "A",
		SomeCount: 5,
	})

	index.Update(&TestStruct{
		ID:        "2",
		Name:      "A",
		SomeCount: 7,
	})
	index.Update(&TestStruct{
		ID:        "3",
		Name:      "B",
		SomeCount: 7,
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().Is("Name", "A").Is("SomeCount", 7).List())
	if res != "2" {
		t.Fatal(res)
	}

}

func TestTags(t *testing.T) {
	index := prepareTestIndex()
	index.Update(&TestStruct{
		ID:   "1",
		Tags: []string{"hello", "world"},
	})

	index.Update(&TestStruct{
		ID:   "2",
		Tags: []string{"apple", "pear"},
	})
	index.Update(&TestStruct{
		ID:   "3",
		Tags: []string{"one", "two"},
	})

	index.Flush()
	index.Refresh()

	for k, v := range [][]string{
		{"1", "hello"},
		{"2", "apple"},
	} {
		res := getIDS(index.Query().Is("Tags", v[0:]).List())
		if res != v[0] {
			t.Fatal(k, res)
		}
	}
}

func TestCzechSearch(t *testing.T) {
	index := prepareTestIndex()
	index.Update(&TestStruct{
		ID:   "1",
		Text: "Nový náměstek ministra baobab průmyslu auto se sešel s Topolánkem, který pracuje pro Křetínského. „Příště si dám pozor,“ říká",
	})

	index.Update(&TestStruct{
		ID:   "2",
		Text: "Padají i konkrétní jména.",
	})

	index.Flush()
	index.Refresh()

	for k, v := range [][2]string{
		{"baobab", "1"},
		{"jméno", "2"},
		{"priste", "1"},
	} {
		res := getIDS(index.Query().Is("Text", v[0]).List())
		if res != v[1] {
			t.Fatal(k, res)
		}
	}

}

func TestAllQuery(t *testing.T) {
	index := prepareTestIndex()
	index.Update(&TestStruct{
		ID:        "1",
		Name:      "C",
		SomeCount: 1,
	})

	index.Update(&TestStruct{
		ID:        "2",
		Name:      "A",
		SomeCount: 3,
		IsOK:      true,
	})
	index.Update(&TestStruct{
		ID:        "3",
		Name:      "B",
		SomeCount: 2,
	})

	index.Flush()
	index.Refresh()

	expected := getIDS(index.Query().Sort("ID", false).List())
	if expected != "3,2,1" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", true).List())
	if expected != "2,3,1" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).List())
	if expected != "1,3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).Limit(2).List())
	if expected != "1,3" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).Limit(2).Offset(1).List())
	if expected != "3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("SomeCount", true).List())
	if expected != "1,3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("IsOK", false).Limit(1).List())
	if expected != "2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Is("IsOK", true).List())
	if expected != "2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Is("Name", "B").List())
	if expected != "3" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Is("SomeCount", 3).List())
	if expected != "2" {
		t.Fatal(expected)
	}

}

func TestBasic(t *testing.T) {
	index := prepareTestIndex()

	err := index.Update(&TestStruct{
		ID:   "2",
		Name: "A",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.Update(&TestStruct{
		ID:   "5",
		Name: "B",
	})

	index.Flush()

	c, err := index.Count()
	if err != nil {
		t.Fatal(err)
	}
	if c != 2 {
		t.Fatal("wrong size", c)
	}

	item, _ := index.Get("5")
	if item.Name != "B" {
		t.Fatal(item.Name)
	}

	err = index.DeleteItem("2")
	if err != nil {
		t.Fatal(err)
	}

	index.Flush()

	c, _ = index.Count()
	if c != 1 {
		t.Fatal(c)
	}

	err = index.Delete()
	if err != nil {
		t.Fatal(err)
	}

}
