package administration

import (
	"strings"
	"testing"
	"time"
)

type ResourceStruct struct {
	ID          int64
	Name        string
	Other       string
	Showing     string `prago-preview:"true"`
	IsSomething bool
	Floating    float64
	Date        time.Time `prago-type:"date"`
	Count       int64
	privateint  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func prepareResource() (*Admin, *Resource) {
	admin := NewAdmin("admin", "AAA")
	admin.db = db
	resource, err := admin.CreateResource(ResourceStruct{})
	if err != nil {
		panic(err)
	}
	admin.UnsafeDropTables()
	admin.Migrate(false)
	return admin, resource
}

func TestAdminQuery(t *testing.T) {
	var err error
	var item ResourceStruct
	var createdItem interface{}
	var resource *Resource

	admin, resource := prepareResource()

	admin.UnsafeDropTables()
	admin.Migrate(false)

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

	resource.newItem(&createdItem)
	err = admin.Query().Where(2).Get(createdItem)
	if err != nil {
		t.Fatal(err)
	}
	if createdItem.(*ResourceStruct).Name != "C" {
		t.Fatal(createdItem.(*ResourceStruct).Name)
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

func TestResource(t *testing.T) {
	admin, resource := prepareResource()

	items, err := resource.getListContent(admin, &listRequest{OrderBy: "id"}, &User{})
	if err != nil {
		t.Fatal(err)
	}

	var item interface{}
	resource.newItem(&item)
	count, err := admin.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal(count)
	}

	admin.Create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
	admin.Create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err = admin.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ = resource.getListContent(admin, &listRequest{OrderBy: "id", Page: 1}, &User{})

	t.Log(items)
	t.Log(items.Rows)

	if len(items.Rows[0].Items) != 3 {
		t.Fatal("wrong length")
	}

	if items.Rows[1].Items[2].Value != "show" {
		t.Fatal(items.Rows[1].Items[2].Value)
	}
}

func TestResourceUnique(t *testing.T) {
	type ResourceStructUnique struct {
		ID   int64
		Name string `prago-unique:"true"`
	}

	admin, _ := prepareResource()
	resource, _ := admin.CreateResource(ResourceStructUnique{})
	admin.UnsafeDropTables()
	admin.Migrate(false)

	admin.Create(&ResourceStructUnique{Name: "A"})
	admin.Create(&ResourceStructUnique{Name: "B"})
	admin.Create(&ResourceStructUnique{Name: "A"})

	var item interface{}
	resource.newItem(&item)
	count, err := admin.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}
}

func TestResourceDate(t *testing.T) {
	admin, resource := prepareResource()
	tm := time.Now()

	admin.Create(&ResourceStruct{Date: tm})

	var item interface{}
	resource.newItem(&item)
	err := admin.Query().WhereIs("date", tm.Format("2006-01-02")).Get(item)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResourceTimestamps(t *testing.T) {
	admin, resource := prepareResource()

	testStartTime := time.Now().Truncate(time.Second)

	admin.Create(&ResourceStruct{Name: "A"})

	var itemIface interface{}
	resource.newItem(&itemIface)
	err := admin.Query().WhereIs("id", 1).Get(itemIface)
	if err != nil {
		t.Fatal(err)
	}

	item := itemIface.(*ResourceStruct)

	if item.UpdatedAt.Before(testStartTime) || time.Now().Before(item.UpdatedAt) {
		t.Fatal(item.UpdatedAt)
	}

	if item.CreatedAt.Before(testStartTime) || time.Now().Before(item.CreatedAt) {
		t.Fatal(item.CreatedAt)
	}
}

func TestResourceBool(t *testing.T) {
	admin, resource := prepareResource()

	admin.Create(&ResourceStruct{Name: "A", IsSomething: false})
	admin.Create(&ResourceStruct{Name: "B", IsSomething: true})

	var itemIface interface{}
	resource.newItem(&itemIface)
	err := admin.Query().WhereIs("issomething", true).Get(itemIface)

	if err != nil {
		t.Fatal(err)
	}

	item := itemIface.(*ResourceStruct)
	if item.Name != "B" {
		t.Fatal(item)
	}

	err = admin.Query().WhereIs("issomething", false).Get(itemIface)
	if err != nil {
		t.Fatal(err)
	}

	item = itemIface.(*ResourceStruct)
	if item.Name != "A" {
		t.Fatal(item)
	}
}

func TestResourceCreateWithID(t *testing.T) {
	admin, resource := prepareResource()
	admin.Create(&ResourceStruct{ID: 85, Name: "A"})

	var item interface{}
	resource.newItem(&item)

	admin.Query().Get(item)
	id := item.(*ResourceStruct).ID
	if id != 85 {
		t.Fatal(id)
	}
}

func TestShouldNotSaveWithZeroID(t *testing.T) {
	admin, _ := prepareResource()
	err := admin.Save(&ResourceStruct{})
	if err == nil {
		t.Fatal("should not be nil")
	}

}

func TestShouldNotCreateResourceWithPointer(t *testing.T) {
	var err error
	admin, _ := prepareResource()
	_, err = admin.CreateResource(ResourceStruct{})
	if err == nil {
		t.Fatal("Should have non nil error")
	}
	_, err = admin.CreateResource(&ResourceStruct{})
	if err == nil {
		t.Fatal("Should have non nil error")
	}
	_, err = admin.CreateResource(85)
	if err == nil {
		t.Fatal("Should have non nil error")
	}
}

func TestLongSaveText(t *testing.T) {
	text := "some" + string(make([]byte, 100000))
	admin, _ := prepareResource()
	err := admin.Create(&ResourceStruct{Name: text})
	if err != nil {
		t.Fatal(err)
	}
	var item ResourceStruct
	admin.Query().WhereIs("id", 1).Get(&item)

	if !strings.HasPrefix(item.Name, "some") {
		t.Fatal(item.Name)
	}

}
