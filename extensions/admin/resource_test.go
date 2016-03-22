package admin

import (
	"testing"
	"time"
)

type ResourceStruct struct {
	ID        int64
	Name      string
	Other     string
	Showing   string `prago-admin-show:"yes"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func prepareResource() *AdminResource {
	resource, _ := NewResource(ResourceStruct{})
	resource.admin = dbProvider{}

	resource.UnsafeDropTable()
	resource.Migrate()
	return resource
}

func TestResource(t *testing.T) {
	resource := prepareResource()

	resource.Create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
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
