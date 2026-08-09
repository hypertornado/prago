package prago

import (
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource(t)

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	resource.app.Create(&resA)
	resource.app.Create(&resB)

	item := resource.app.Query[ResourceStruct]().Is("id", resB.ID).First()
	if item == nil {
		t.Fatal("is nil")
	}

	if item.Name != "B" {
		t.Fatal("wrong name")
	}

	list := resource.app.Query[ResourceStruct]().List()
	if len(list) != 2 {
		t.Fatal("wrong length")
	}
}
