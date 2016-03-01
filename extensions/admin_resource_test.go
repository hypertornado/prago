package extensions

import (
	"fmt"
	"testing"
)

type ResourceStruct struct {
	ID   int64
	Name string
}

func TestResource(t *testing.T) {
	resource, _ := NewResource(ResourceStruct{})
	resource.admin = dbProvider{}

	resource.Migrate()
	resource.Create(&ResourceStruct{Name: "First"})
	resource.Create(&ResourceStruct{Name: "Second"})

	count, err := resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ := resource.ListTableItems()

	fmt.Println(items)

}
