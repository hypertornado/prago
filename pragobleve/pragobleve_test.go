package pragobleve

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestStruct struct {
	ID   string
	Name string

	OtherName string

	IsOK bool

	Tags []string

	SomeCount int64
}

func prepareTestIndex() *PragoBleveIndex[TestStruct] {
	pb := New("/Users/odchazel/.pragobleve")
	pb.DeleteIndexByName("teststruct")
	index := NewIndex[TestStruct](pb)
	err := index.Create()
	if err != nil {
		panic(err)
	}
	return index
}

func testSearchResult(t *testing.T, query *PragoBleveQuery[TestStruct], expected []string) {
	res, err := query.Search()
	must(err)
	var ids []string
	for _, v := range res {
		ids = append(ids, v.ID)
	}

	resJSON, err := json.Marshal(ids)
	must(err)

	expectedJSON, err := json.Marshal(expected)
	must(err)

	if string(resJSON) != string(expectedJSON) {
		t.Fatalf("expected: %s, got: %s", string(expectedJSON), string(resJSON))
	}

}

func TestPragobleve(t *testing.T) {
	testIndex := prepareTestIndex()

	if testIndex.Size() != 0 {
		panic(fmt.Sprintf("wrong size: %d", testIndex.Size()))
	}

	must(testIndex.Save(&TestStruct{
		ID:   "1",
		Name: "hello",

		SomeCount: 43,
	}))

	if testIndex.Size() != 1 {
		panic(fmt.Sprintf("wrong size: %d", testIndex.Size()))
	}

	must(testIndex.Save(&TestStruct{
		ID:   "2",
		Name: "world",

		SomeCount: 444,
	}))

	must(testIndex.Save(&TestStruct{
		ID:   "3",
		Name: "j doiwq jdwioq hello world",
	}))

	doc := testIndex.Get("2")
	if doc.Name != "world" {
		t.Fatal(doc.Name)
	}
	if doc.SomeCount != 444 {
		t.Fatal(doc.SomeCount)
	}

}

func TestPhrase(t *testing.T) {
	testIndex := prepareTestIndex()

	must(testIndex.Save(&TestStruct{
		ID:   "1",
		Name: "id wjq dowqji",
	}))

	must(testIndex.Save(&TestStruct{
		ID:   "2",
		Name: "world",
	}))

	must(testIndex.Save(&TestStruct{
		ID:   "3",
		Name: "karel hello world",
	}))

	testIndex.Query().Phrase("Name", "hello")

	testSearchResult(t, testIndex.Query().Phrase("Name", "hello"), []string{"3"})
}

func TestBoolean(t *testing.T) {
	testIndex := prepareTestIndex()

	must(testIndex.Save(&TestStruct{
		ID:   "1",
		IsOK: false,
	}))

	must(testIndex.Save(&TestStruct{
		ID:   "2",
		IsOK: true,
	}))
	testSearchResult(t, testIndex.Query().Bool("IsOK", true), []string{"2"})
}

func TestSort(t *testing.T) {
	testIndex := prepareTestIndex()
	must(testIndex.Save(&TestStruct{
		ID:        "1",
		SomeCount: 10,
	}))

	must(testIndex.Save(&TestStruct{
		ID:        "2",
		SomeCount: 5,
	}))
	must(testIndex.Save(&TestStruct{
		ID:        "3",
		SomeCount: 20,
	}))
	testSearchResult(t, testIndex.Query().Sort("SomeCount"), []string{"2", "1", "3"})

}

func TestTags(t *testing.T) {
	testIndex := prepareTestIndex()
	must(testIndex.Save(&TestStruct{
		ID:   "1",
		Tags: []string{"hello"},
	}))

	must(testIndex.Save(&TestStruct{
		ID:   "2",
		Tags: []string{"its", "beautiful", "world", "and", "i"},
	}))
	must(testIndex.Save(&TestStruct{
		ID:   "3",
		Tags: []string{"its", "worlds"},
	}))
	testSearchResult(t, testIndex.Query().Phrase("Tags", "world"), []string{"2"})

}

func TestOffset(t *testing.T) {
	testIndex := prepareTestIndex()
	must(testIndex.Save(&TestStruct{
		ID:        "1",
		SomeCount: 2,
	}))

	must(testIndex.Save(&TestStruct{
		ID:        "2",
		SomeCount: 4,
	}))
	must(testIndex.Save(&TestStruct{
		ID:        "3",
		SomeCount: 6,
	}))
	testSearchResult(t, testIndex.Query().Sort("SomeCount").Offset(1), []string{"2", "3"})

	testSearchResult(t, testIndex.Query().Sort("SomeCount").Offset(1).Size(1), []string{"2"})

}
