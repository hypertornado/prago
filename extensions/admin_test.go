package extensions

import (
	"testing"
)

func prepareAdmin() *Admin {
	admin := NewAdmin("admin", "AAA")
	admin.db = db
	return admin
}

func TestAdmin(t *testing.T) {
	admin := prepareAdmin()
	admin.CreateResources(ResourceStruct{})
	admin.Migrate()

	var err error

	err = admin.Create(&ResourceStruct{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}

	admin.Create(&ResourceStruct{Name: "B"})

	var item ResourceStruct
	admin.Query().Get(&item)

	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	var list []*ResourceStruct
	err = admin.Query().Get(&list)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatal(len(list))
	}

	if list[1].Name != "B" {
		t.Fatal(list[1].Name)
	}
}
