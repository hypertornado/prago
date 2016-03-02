package extensions

import (
	"fmt"
	"testing"
)

type ResourceStruct struct {
	ID      int64
	Name    string
	Other   string
	Showing string `prago-admin-show:"yes"`
}

func TestResource(t *testing.T) {
	resource, _ := NewResource(ResourceStruct{})
	resource.admin = dbProvider{}

	err := resource.Migrate()
	if err != nil {
		t.Fatal(err)
	}
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
