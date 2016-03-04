package extensions

import (
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
	resource.Create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err := resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ := resource.ListTableItems()

	if len(items.Header) != 3 {
		t.Fatal(len(items.Header))
	}

	if items.Rows[1].Items[2].Value.(string) != "show" {
		t.Fatal(items.Rows[1].Items[2].Value.(string))
	}
}
