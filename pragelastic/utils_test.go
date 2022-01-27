package pragelastic

import "testing"

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

}
