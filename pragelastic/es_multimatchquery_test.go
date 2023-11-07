package pragelastic

import (
	"testing"
)

func TestMultimatchQuery(t *testing.T) {
	index := prepareTestIndex[TestStruct]()

	index.UpdateSingle(&TestStruct{
		ID:          "1",
		Description: "",
		Text:        "",
	})

	index.UpdateSingle(&TestStruct{
		ID:          "2",
		Description: "hello",
		Text:        "something else",
	})
	index.UpdateSingle(&TestStruct{
		ID:          "3",
		Description: "hello hello",
		Text:        "hello hello",
	})
	index.UpdateSingle(&TestStruct{
		ID:          "4",
		Description: "",
		Text:        "hello",
	})

	index.Flush()
	index.Refresh()

	qs := "hello"
	mq := NewMultiMatchQuery(qs)
	mq.FieldWithBoost("Description", 3)
	mq.FieldWithBoost("Text", 1)
	mq.Type("most_fields")
	mq.MinimumShouldMatch("0")

	ids := getIDS(index.Query().ShouldQuery(mq).mustList())
	if ids != "3,2,4" {
		t.Fatal(ids)
	}

}
