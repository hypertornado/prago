package pragelastic

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const testClientName = "pragelastic-test"

type TestStruct struct {
	ID              string
	Name            string `elastic-datatype:"keyword"`
	Text            string `elastic-datatype:"text" elastic-analyzer:"czech"`
	SomeCount       int64
	Score           float64
	NonIndexedField string `elastic-datatype:"text" elastic-enabled:"false"`
	IsOK            bool
	Time            time.Time
	Tags            []string

	KVMap map[string]float64

	MapAr map[int64][]int64
}

func getIDS[T any](items []*T) string {
	ret := []string{}
	for _, v := range items {
		id := reflect.ValueOf(v).Elem().FieldByName("ID").String()
		ret = append(ret, id)
	}
	return strings.Join(ret, ",")
}

func prepareTestIndex[T any]() *Index[T] {
	lib, err := New(testClientName)
	index := NewIndex[T](lib)
	index.Delete()
	err = index.Create()
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
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		Name:      "A",
		SomeCount: 5,
	})

	index.UpdateSingle(&TestStruct{
		ID:        "2",
		Name:      "A",
		SomeCount: 7,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		Name:      "B",
		SomeCount: 7,
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().Filter("Name", "A").Filter("SomeCount", 7).mustList())
	if res != "2" {
		t.Fatal(res)
	}

}

func TestMap(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID: "1",
		KVMap: map[string]float64{
			"cs": 11,
			"en": 23,
		},

		MapAr: map[int64][]int64{
			32: {11, 12},
		},
	})

	index.UpdateSingle(&TestStruct{
		ID: "2",
		KVMap: map[string]float64{
			"cs": 21,
			"en": 21,
		},
		MapAr: map[int64][]int64{
			32: {12},
			33: {11, 12},
		},
	})
	index.UpdateSingle(&TestStruct{
		ID: "3",
		KVMap: map[string]float64{
			"cs": 31,
			"en": 33,
		},
		MapAr: map[int64][]int64{
			32: {11},
		},
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().Sort("KVMap.en", true).mustList())
	if res != "2,1,3" {
		t.Fatal(res)
	}

	res = getIDS(index.Query().Filter("KVMap.en", 21).mustList())
	if res != "2" {
		t.Fatal(res)
	}

	_, _, err := index.Query().Filter("MapAr.32", 13).List()
	if err != nil {
		t.Fatal(err)
	}

	res = getIDS(index.Query().Filter("MapAr.32", 11).mustList())
	if res != "1,3" {
		t.Fatal(res)
	}
}

func TestNonIndexedField(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:              "1",
		NonIndexedField: "hello world",
		SomeCount:       5,
	})
	index.Flush()
	index.Refresh()

	items, _, _ := index.Query().Filter("NonIndexedField", "hello").List()
	if len(items) != 0 {
		t.Fatal("expected 0 results for querying diabled field")
	}

}

func TestTags(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:   "1",
		Tags: []string{"hello", "world"},
	})

	index.UpdateSingle(&TestStruct{
		ID:   "2",
		Tags: []string{"apple", "pear"},
	})
	index.UpdateSingle(&TestStruct{
		ID:   "3",
		Tags: []string{"one", "two"},
	})

	index.Flush()
	index.Refresh()

	for k, v := range [][]string{
		{"1", "hello"},
		{"2", "apple"},
	} {
		res := getIDS(index.Query().Filter("Tags", v[0:]).mustList())
		if res != v[0] {
			t.Fatal(k, res)
		}
	}
}

func TestKeywordFilter(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:   "1",
		Name: "LIB",
	})
	index.UpdateSingle(&TestStruct{
		ID:   "2",
		Name: "TLP",
	})
	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().Filter("Name", "LIB").mustList())
	if res != "1" {
		t.Fatal(res)
	}

}

func TestDeleteQuery(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:   "1",
		Name: "LIB",
	})
	index.UpdateSingle(&TestStruct{
		ID:   "2",
		Name: "TLP",
	})
	index.UpdateSingle(&TestStruct{
		ID:   "3",
		Name: "TLP",
	})
	index.Flush()
	index.Refresh()

	err := index.Query().Filter("Name", "TLP").Delete()
	if err != nil {
		t.Fatal(err)
	}

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().mustList())
	if res != "1" {
		t.Fatal(res)
	}

}

func TestCzechSearch(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:   "1",
		Text: "Nový náměstek ministra baobab průmyslu auto se sešel s Topolánkem, který pracuje pro Křetínského. „Příště si dám pozor,“ říká",
	})

	index.UpdateSingle(&TestStruct{
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
		res := getIDS(index.Query().Filter("Text", v[0]).mustList())
		if res != v[1] {
			t.Fatal(k, res)
		}
	}

}

func TestAllQuery(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		Name:      "C",
		Score:     2.34,
		Time:      time.Date(2020, 10, 10, 5, 1, 1, 1, time.UTC),
		SomeCount: 1,
	})

	index.UpdateSingle(&TestStruct{
		ID:        "2",
		Name:      "A",
		Score:     2.1,
		SomeCount: 3,
		Time:      time.Date(2019, 10, 10, 5, 1, 0, 1, time.UTC),
		IsOK:      true,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		Name:      "B",
		Score:     2.8,
		Time:      time.Date(2019, 10, 10, 5, 1, 1, 1, time.UTC),
		SomeCount: 2,
	})

	index.Flush()
	index.Refresh()

	expected := getIDS(index.Query().Sort("ID", false).mustList())
	if expected != "3,2,1" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", true).mustList())
	if expected != "2,3,1" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).mustList())
	if expected != "1,3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Time", false).mustList())
	if expected != "1,3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).Limit(2).mustList())
	if expected != "1,3" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Name", false).Limit(2).Offset(1).mustList())
	if expected != "3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("SomeCount", true).mustList())
	if expected != "1,3,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("IsOK", false).Limit(1).mustList())
	if expected != "2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Score", true).mustList())
	if expected != "2,1,3" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Sort("Score", false).mustList())
	if expected != "3,1,2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Filter("IsOK", true).mustList())
	if expected != "2" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Filter("Name", "B").mustList())
	if expected != "3" {
		t.Fatal(expected)
	}

	expected = getIDS(index.Query().Filter("SomeCount", 3).mustList())
	if expected != "2" {
		t.Fatal(expected)
	}

}

func TestBasic(t *testing.T) {
	index := prepareTestIndex[TestStruct]()

	err := index.UpdateSingle(&TestStruct{
		ID:   "2",
		Name: "A",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = index.UpdateSingle(&TestStruct{
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
