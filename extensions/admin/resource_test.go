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
	privateint  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func prepareResource() *Resource {
	resource, _ := newResource(ResourceStruct{})
	resource.admin = dbTestProvider{}

	resource.unsafeDropTable()
	resource.migrate(false)
	return resource
}

func TestResource(t *testing.T) {
	resource := prepareResource()

	items, err := resource.getList("en", "", make(map[string][]string))
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

	resource.create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
	resource.create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err = resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ = resource.getList("en", "", make(map[string][]string))

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

	resource, _ := newResource(ResourceStructUnique{})
	resource.admin = dbTestProvider{}

	resource.unsafeDropTable()
	resource.migrate(false)

	resource.create(&ResourceStructUnique{Name: "A"})
	resource.create(&ResourceStructUnique{Name: "B"})
	resource.create(&ResourceStructUnique{Name: "A"})

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

	resource.create(&ResourceStruct{Date: tm})
	_, err := resource.Query().Where(map[string]interface{}{"date": tm.Format("2006-01-02")}).First()
	if err != nil {
		t.Fatal(err)
	}

}

func TestResourceTimestamps(t *testing.T) {
	resource := prepareResource()

	testStartTime := time.Now().Truncate(time.Second)

	resource.create(&ResourceStruct{Name: "A"})

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

	resource.create(&ResourceStruct{Name: "A", IsSomething: false})
	resource.create(&ResourceStruct{Name: "B", IsSomething: true})

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
	resource.create(&ResourceStruct{ID: 85, Name: "A"})
	item, _ := resource.Query().First()
	id := item.(*ResourceStruct).ID
	if id != 85 {
		t.Fatal(id)
	}
}

func TestShouldNotSaveWithZeroID(t *testing.T) {
	resource := prepareResource()
	err := resource.save(&ResourceStruct{})
	if err == nil {
		t.Fatal("should not be nil")
	}

}

func TestShouldNotCreateResourceWithPointer(t *testing.T) {
	var err error
	_, err = newResource(&ResourceStruct{})
	if err == nil {
		t.Fatal("Should have non nil error")
	}
	_, err = newResource(85)
	if err == nil {
		t.Fatal("Should have non nil error")
	}
}
