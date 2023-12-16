package prago

import (
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource()

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	CreateItem(resource.data.app, &resA)
	CreateItem(resource.data.app, &resB)

	item := Query[ResourceStruct](resource.data.app).Is("id", resB.ID).First()
	if item == nil {
		t.Fatal("is nil")
	}

	if item.Name != "B" {
		t.Fatal("wrong name")
	}

	list := Query[ResourceStruct](resource.data.app).List()
	if len(list) != 2 {
		t.Fatal("wrong length")
	}
}
