package pragelastic

import (
	"testing"
	"time"
)

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

func TestQueryRangeDates(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:   "1",
		Time: time.Date(2022, 2, 21, 0, 0, 0, 0, time.UTC),
	})

	index.UpdateSingle(&TestStruct{
		ID:   "2",
		Time: time.Date(2022, 2, 22, 0, 0, 0, 0, time.UTC),
	})
	index.UpdateSingle(&TestStruct{
		ID:   "3",
		Time: time.Date(2022, 2, 20, 0, 0, 0, 0, time.UTC),
	})

	index.Flush()
	index.Refresh()

	res := getIDS(index.Query().GreaterThanOrEqual("Time",
		time.Date(2022, 2, 21, 0, 0, 0, 0, time.UTC),
	).mustList())
	if res != "1,2" {
		t.Fatal(res)
	}

}
