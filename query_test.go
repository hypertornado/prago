package prago

import (
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource()

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	resource.Create(&resA)
	resource.Create(&resB)

	item := resource.Query().Is("id", resB.ID).First()
	if item == nil {
		t.Fatal("is nil")
	}

	if item.Name != "B" {
		t.Fatal("wrong name")
	}

	list := resource.Query().List()
	if len(list) != 2 {
		t.Fatal("wrong length")
	}

}
