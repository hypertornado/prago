package pragelastic

import (
	"testing"
)

type TestStruct struct {
	ID        string
	Name      string
	SomeCount int64
}

func TestUtils(t *testing.T) {
	id := getID(&TestStruct{
		ID: "85",
	})
	if id != "85" {
		t.Fatal(id)
	}

	fields := getFields[TestStruct]()
	if fields[1].Name != "Name" {
		t.Fatal("wrong")
	}

	lib := New("pragelastic-test")
	index := NewIndex[TestStruct](lib)

	//fmt.Println(index.indexName())

	index.Delete()

	err := index.Create()
	if err != nil {
		t.Fatal(err)
	}

	err = index.Flush()
	if err != nil {
		t.Fatal(err)
	}

	err = index.Update(&TestStruct{
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

	c, err := index.count()
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

	c, _ = index.count()
	if c != 1 {
		t.Fatal(c)
	}

	err = index.Delete()
	if err != nil {
		t.Fatal(err)
	}

}
