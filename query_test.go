package prago

import (
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource()

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	CreateItem(resource.app, &resA)
	CreateItem(resource.app, &resB)

	item := Query[ResourceStruct](resource.app).Is("id", resB.ID).First()
	if item == nil {
		t.Fatal("is nil")
	}

	if item.Name != "B" {
		t.Fatal("wrong name")
	}

	list := Query[ResourceStruct](resource.app).List()
	if len(list) != 2 {
		t.Fatal("wrong length")
	}
}
