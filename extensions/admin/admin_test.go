package admin

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
	var item ResourceStruct

	err = admin.Query().Get(&item)

	err = admin.Create(&ResourceStruct{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}

	admin.Create(&ResourceStruct{Name: "C"})
	admin.Create(&ResourceStruct{Name: "B"})

	admin.Query().Get(&item)

	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	var list []*ResourceStruct
	err = admin.Query().Get(&list)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 3 {
		t.Fatal(len(list))
	}

	if list[2].Name != "B" {
		t.Fatal(list[2].Name)
	}

	count, err := admin.Query().Count(&ResourceStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal(count)
	}

	admin.Query().Limit(1).Offset(1).Limit(1).Get(&list)
	if len(list) != 1 {
		t.Fatal(len(list))
	}
	if list[0].Name != "C" {
		t.Fatal(list[0].Name)
	}
}
