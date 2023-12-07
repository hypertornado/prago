package prago

import (
	"context"
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource()

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	resource.Create(context.Background(), &resA)
	resource.Create(context.Background(), &resB)

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
