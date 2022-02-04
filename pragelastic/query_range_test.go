package pragelastic

import "testing"

func TestQueryRange(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		SomeCount: 1,
	})

	index.UpdateSingle(&TestStruct{
		ID:        "2",
		SomeCount: 2,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		SomeCount: 3,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "4",
		SomeCount: 4,
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().Range("SomeCount", 2, 3).mustList())
	if res != "2,3" {
		t.Fatal(res)
	}

}
