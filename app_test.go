package prago

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

/*
for testing, create database and user in mysql as root

CREATE DATABASE prago_test CHARACTER SET utf8 DEFAULT COLLATE utf8_unicode_ci;
CREATE USER 'prago_test'@'localhost' IDENTIFIED BY 'prago_test';
GRANT ALL ON prago_test.* TO 'prago_test'@'localhost';
FLUSH PRIVILEGES;

*/

type ResourceStruct struct {
	ID          int64
	Name        string `prago-validations:"nonempty"`
	Text        string `prago-type:"text"`
	Other       string
	Showing     string
	IsSomething bool
	Floating    float64
	Date        time.Time `prago-type:"date"`
	Count       int64
	privateint  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ResourceStructUnique struct {
	ID         int64
	UniqueName string `prago-unique:"true"`
}

func prepareResource(t *testing.T) *Resource {
	var resource *Resource
	_ = NewTesting(t, func(app *App) {
		resource = NewResource[ResourceStruct](app)
		NewResource[ResourceStructUnique](app)
	})
	return resource
}

func TestBasicResource2(t *testing.T) {
	resource := prepareResource(t)

	item := &ResourceStruct{Name: "A", Floating: 3.14}

	err := CreateItemWithContext(context.Background(), resource.app, item)
	if err != nil {
		t.Fatal(err)
	}

	if item.ID <= 0 {
		t.Fatal("should be positive")
	}

	item2 := Query[ResourceStruct](resource.app).ID(item.ID)
	if item2 == nil {
		t.Fatal("should not be nil")
	}

	CreateItemWithContext(context.Background(), resource.app, &ResourceStruct{Name: "C"})
	CreateItemWithContext(context.Background(), resource.app, &ResourceStruct{Name: "B"})

	list := Query[ResourceStruct](resource.app).List()
	if len(list) != 3 {
		t.Fatalf("wrong length %d", len(list))
	}

	first := Query[ResourceStruct](resource.app).Is("id", item.ID).First()
	if first.Name != "A" {
		t.Fatal("wrong name")
	}

	if Query[ResourceStruct](resource.app).Is("id", item.ID).First().Name != "A" {
		t.Fatal("wrong name")
	}

	item.Name = "changed"

	err = UpdateItem(resource.app, item)
	if err != nil {
		t.Fatal(err)
	}

	if Query[ResourceStruct](resource.app).Is("id", item.ID).First().Name != "changed" {
		t.Fatal("wrong name")
	}

	first = Query[ResourceStruct](resource.app).Is("name", "B").First()
	if first.Name != "B" {
		t.Fatal("wrong name")
	}

	count, _ := Query[ResourceStruct](resource.app).Count()
	if count != 3 {
		t.Fatalf("wrong count %d", count)
	}

	err = DeleteItem[ResourceStruct](resource.app, item.ID)
	if err != nil {
		t.Fatal(err)
	}

	count, _ = Query[ResourceStruct](resource.app).Count()
	if count != 2 {
		t.Fatalf("wrong count %d", count)
	}

}

func TestQuery(t *testing.T) {
	resource := prepareResource(t)

	err := CreateItem(resource.app, &ResourceStruct{Name: "A", Floating: 3.14})
	if err != nil {
		t.Fatal(err)
	}
	CreateItem(resource.app, &ResourceStruct{Name: "C"})
	CreateItem(resource.app, &ResourceStruct{Name: "B"})

	item := Query[ResourceStruct](resource.app).Where("id = ?", 2).First()
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	createdItem := Query[ResourceStruct](resource.app).Where("id = ?", 2).First()
	if createdItem == nil {
		t.Fatal("should not be nil")
	}
	if createdItem.Name != "C" {
		t.Fatal(createdItem.Name)
	}

	item = Query[ResourceStruct](resource.app).Where("id=?", 2).First()
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	item = Query[ResourceStruct](resource.app).First()
	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	if item.Floating < 3 || item.Floating > 4 {
		t.Fatal(item.Floating)
	}

	var list []*ResourceStruct
	list = Query[ResourceStruct](resource.app).List()
	if len(list) != 3 {
		t.Fatal(len(list))
	}

	if list[2].Name != "B" {
		t.Fatal(list[2].Name)
	}

	count, err := Query[ResourceStruct](resource.app).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal(count)
	}

	list = Query[ResourceStruct](resource.app).Limit(1).Offset(1).Limit(1).List()
	if len(list) != 1 {
		t.Fatal(len(list))
	}
	if list[0].Name != "C" {
		t.Fatal(list[0].Name)
	}

	err = DeleteItem[ResourceStruct](resource.app, item.ID)
	if err != nil {
		t.Fatal(err)
	}

	count, _ = Query[ResourceStruct](resource.app).Count()
	if count != 2 {
		t.Fatal(count)
	}
}

func TestQueryIn(t *testing.T) {

	resource := prepareResource(t)

	resources := []*ResourceStruct{
		&ResourceStruct{
			Name: "a",
		},
		&ResourceStruct{
			Name: "b",
		},
		&ResourceStruct{
			Name: "c",
		},
	}
	for _, v := range resources {
		CreateItem(resource.app, v)
	}

	items := Query[ResourceStruct](resource.app).In("id", []int64{resources[0].ID, resources[1].ID}).Order("id").List()
	if len(items) != 2 {
		t.Fatal(items)
	}
	if items[0].ID != resources[0].ID {
		t.Fatal(items[0])
	}

	items = Query[ResourceStruct](resource.app).In("id", fmt.Sprintf(";%d;%d;", resources[0].ID, resources[1].ID)).Order("id").List()
	if len(items) != 2 {
		t.Fatal(items)
	}

	items = Query[ResourceStruct](resource.app).In("id", resources[0].ID).Order("id").List()
	if len(items) != 1 {
		t.Fatal(items)
	}

}

func TestResource(t *testing.T) {
	resource := prepareResource(t)
	items, err := resource.getListContent(context.Background(), resource.app.newUserData(&user{Role: "sysadmin"}), map[string][]string{
		"_order": {"id"},
	})
	if err != nil {
		t.Fatal(err)
	}

	count, err := Query[ResourceStruct](resource.app).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal(count)
	}

	err = CreateItem(resource.app, &ResourceStruct{Name: "First", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	CreateItem(resource.app, &ResourceStruct{Name: "Second", Showing: "show"})

	count, err = Query[ResourceStruct](resource.app).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal(count)
	}

	items, err = resource.getListContent(context.Background(), resource.app.newUserData(&user{Role: "sysadmin"}), map[string][]string{
		"_order": {"id"},
		"_page":  {"1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(items.Rows[0].Items) != 11 {
		t.Fatalf("wrong length: %d", len(items.Rows[0].Items))
	}

	if items.Rows[1].Items[4].Name != "show" {
		t.Fatal(items.Rows[1].Items[4])
	}
}

func TestResourceUnique(t *testing.T) {
	app := prepareResource(t).app

	resource := getResource[ResourceStructUnique](app)

	must(CreateItem(resource.app, &ResourceStructUnique{UniqueName: "A"}))
	must(CreateItem(resource.app, &ResourceStructUnique{UniqueName: "B"}))
	err := CreateItem(resource.app, &ResourceStructUnique{UniqueName: "A"})
	if err == nil {
		t.Fatal("Should fail")
	}

	count, err := Query[ResourceStructUnique](resource.app).Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}
}

func TestResourceDate(t *testing.T) {
	resource := prepareResource(t)
	tm := time.Now()

	CreateItem(resource.app, &ResourceStruct{Date: tm})

	first := Query[ResourceStruct](resource.app).Is("date", tm.Format("2006-01-02")).First()
	if first == nil {
		t.Fatal("should not be nil")
	}
}

func TestResourceTimestamps(t *testing.T) {
	resource := prepareResource(t)

	testStartTime := time.Now().Truncate(time.Second)

	CreateItem(resource.app, &ResourceStruct{Name: "A"})

	item := Query[ResourceStruct](resource.app).Is("id", 1).First()

	if item.UpdatedAt.Before(testStartTime) || time.Now().Before(item.UpdatedAt) {
		t.Fatal(item.UpdatedAt)
	}

	if item.CreatedAt.Before(testStartTime) || time.Now().Before(item.CreatedAt) {
		t.Fatal(item.CreatedAt)
	}
}

func TestResourceBool(t *testing.T) {
	resource := prepareResource(t)

	CreateItem(resource.app, &ResourceStruct{Name: "A", IsSomething: false})
	CreateItem(resource.app, &ResourceStruct{Name: "B", IsSomething: true})

	trueItem := Query[ResourceStruct](resource.app).Is("issomething", true).First()
	if trueItem.Name != "B" {
		t.Fatal(trueItem.Name)
	}

	falseItem := Query[ResourceStruct](resource.app).Is("issomething", false).First()
	if falseItem.Name != "A" {
		t.Fatal(trueItem.Name)
	}
}

func TestResourceCreateWithID(t *testing.T) {
	resource := prepareResource(t)
	CreateItem(resource.app, &ResourceStruct{ID: 85, Name: "A"})

	item := Query[ResourceStruct](resource.app).First()
	id := item.ID
	if id != 85 {
		t.Fatal(id)
	}
}

func TestShouldNotSaveWithZeroID(t *testing.T) {
	resource := prepareResource(t)
	err := UpdateItem(resource.app, &ResourceStruct{})
	if err == nil {
		t.Fatal("should not be nil")
	}
}

func TestWorkingWithConcreteID(t *testing.T) {
	resource := prepareResource(t)
	item := &ResourceStruct{
		ID:   3,
		Name: "A",
	}
	err := CreateItem(resource.app, item)
	if err != nil {
		t.Fatal(err)
	}

	err = DeleteItem[ResourceStruct](resource.app, 3)
	if err != nil {
		t.Fatal(err)
	}

	item.Name = "B"

	err = CreateItem(resource.app, item)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReplace(t *testing.T) {
	resource := prepareResource(t)
	var id int64 = 3
	item := &ResourceStruct{
		ID:   id,
		Name: "A",
	}
	err := Replace(context.Background(), resource.app, item)
	if err != nil {
		t.Fatal(err)
	}
	if Query[ResourceStruct](resource.app).Is("id", id).First() == nil {
		t.Fatal("should not be nil")
	}
	item.Name = "B"
	err = Replace(context.Background(), resource.app, item)
	if err != nil {
		t.Fatal(err)
	}

	count, _ := Query[ResourceStruct](resource.app).Count()
	if count != 1 {
		t.Fatal(count)
	}

	modified := Query[ResourceStruct](resource.app).Is("id", id).First()
	if modified.Name != "B" {
		t.Fatal(modified.Name)
	}
}

func TestLongSaveText(t *testing.T) {
	//TODO: make it work with 100000
	text := "some" + string(make([]byte, 10000))
	resource := prepareResource(t)
	newItem := &ResourceStruct{Text: text}
	err := CreateItem(resource.app, newItem)
	if err != nil {
		t.Fatal(err)
	}
	item := Query[ResourceStruct](resource.app).Is("id", newItem.ID).First()

	if !strings.HasPrefix(item.Text, "some") {
		t.Fatal(item.Text)
	}
}
