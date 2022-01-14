package prago

import (
	"reflect"
	"testing"
)

func TestReflectQuery(t *testing.T) {
	resource := prepareResource()

	resA := ResourceStruct{Name: "A"}
	resB := ResourceStruct{Name: "B"}

	resource.Create(&resA)
	resource.Create(&resB)

	item, err := resource.Query().Is("id", resB.ID).query.first()
	if err != nil {
		t.Fatal(err)
	}

	resStruct, ok := item.(*ResourceStruct)
	if !ok {
		t.Fatalf("wrong type: %s", reflect.TypeOf(item))
	}
	if resStruct.Name != "B" {
		t.Fatal("wrong name")
	}

	list, err := resource.Query().query.list()
	if err != nil {
		t.Fatal(err)
	}
	listStruct, ok := list.([]*ResourceStruct)
	if !ok {
		t.Fatalf("wrong type: %s", reflect.TypeOf(list))
	}
	if len(listStruct) != 2 {
		t.Fatal("wrong length")
	}

}
