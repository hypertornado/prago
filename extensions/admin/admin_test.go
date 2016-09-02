package admin

import (
	"testing"
)

func prepareAdmin() *Admin {
	admin := NewAdmin("admin", "AAA")
	admin.db = db
	return admin
}

func TestAdminQuery(t *testing.T) {
	admin := prepareAdmin()
	admin.CreateResource(ResourceStruct{})
	admin.UnsafeDropTables()
	admin.Migrate(false)

	var err error
	var item ResourceStruct

	err = admin.Create(&ResourceStruct{Name: "A", Floating: 3.14})
	if err != nil {
		t.Fatal(err)
	}

	admin.Create(&ResourceStruct{Name: "C"})
	admin.Create(&ResourceStruct{Name: "B"})

	err = admin.Query().Where(2).Get(&item)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	err = admin.Query().Where("id=?", 2).Get(&item)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	admin.Query().Get(&item)
	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	if item.Floating < 3 || item.Floating > 4 {
		t.Fatal(item.Floating)
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

	if count, _ = admin.Query().WhereIs("name", "A").Delete(&ResourceStruct{}); count != 1 {
		t.Fatal(count)
	}

	if count, _ = admin.Query().Count(&ResourceStruct{}); count != 2 {
		t.Fatal(count)
	}

}
