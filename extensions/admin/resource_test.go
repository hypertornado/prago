package admin

import (
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
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func prepareResource() *AdminResource {
	resource, _ := NewResource(ResourceStruct{})
	resource.admin = dbProvider{}

	resource.UnsafeDropTable()
	resource.migrate(false)
	return resource
}

func TestResource(t *testing.T) {
	resource := prepareResource()

	items, err := resource.GetList("en", "", make(map[string][]string))
	if err != nil {
		t.Fatal(err)
	}

	count, err := resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal(count)
	}

	resource.Create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
	resource.Create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err = resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ = resource.GetList("en", "", make(map[string][]string))

	if len(items.Header) != 3 {
		t.Fatal(len(items.Header))
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

	resource, _ := NewResource(ResourceStructUnique{})
	resource.admin = dbProvider{}

	resource.UnsafeDropTable()
	resource.migrate(false)

	resource.Create(&ResourceStructUnique{Name: "A"})
	resource.Create(&ResourceStructUnique{Name: "B"})
	resource.Create(&ResourceStructUnique{Name: "A"})

	count, err := resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}
}

func TestResourceDate(t *testing.T) {
	resource := prepareResource()
	tm := time.Now()

	resource.Create(&ResourceStruct{Date: tm})
	_, err := resource.Query().Where(map[string]interface{}{"date": tm.Format("2006-01-02")}).First()
	if err != nil {
		t.Fatal(err)
	}

}

func TestResourceTimestamps(t *testing.T) {
	resource := prepareResource()

	testStartTime := time.Now().Truncate(time.Second)

	resource.Create(&ResourceStruct{Name: "A"})

	itemIface, err := resource.Query().Where(map[string]interface{}{"id": 1}).First()
	if err != nil {
		panic(err)
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
	resource := prepareResource()

	resource.Create(&ResourceStruct{Name: "A", IsSomething: false})
	resource.Create(&ResourceStruct{Name: "B", IsSomething: true})

	itemIface, err := resource.Query().Where(map[string]interface{}{"issomething": true}).First()
	if err != nil {
		panic(err)
	}

	item := itemIface.(*ResourceStruct)
	if item.Name != "B" {
		t.Fatal(item)
	}

	itemIface, err = resource.Query().Where(map[string]interface{}{"issomething": false}).First()
	if err != nil {
		panic(err)
	}

	item = itemIface.(*ResourceStruct)
	if item.Name != "A" {
		t.Fatal(item)
	}
}

func TestResourceCreateWithID(t *testing.T) {
	resource := prepareResource()
	resource.Create(&ResourceStruct{ID: 85, Name: "A"})
	item, _ := resource.Query().First()
	id := item.(*ResourceStruct).ID
	if id != 85 {
		t.Fatal(id)
	}

}
