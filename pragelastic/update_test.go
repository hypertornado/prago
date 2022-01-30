package pragelastic

import (
	"fmt"
	"testing"
)

func TestUpdateBulk(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	bu, err := index.UpdateBulk()
	if err != nil {
		t.Fatal(err)
	}
	maxCount := 100
	for i := 0; i < maxCount; i++ {
		bu.AddItem(&TestStruct{
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("A %d", i),
		})
	}

	//other round to make sure rewriting works
	for i := 0; i < maxCount; i++ {
		bu.AddItem(&TestStruct{
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("A %d", i),
		})
	}
	err = bu.Close()
	if err != nil {
		t.Fatal(err)
	}
	bu.index.Flush()
	bu.index.Refresh()

	count, _ := index.Count()
	if count != int64(maxCount) {
		t.Fatal(count)
	}

}
