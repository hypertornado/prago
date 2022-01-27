package prago

import (
	"strings"
	"testing"
	"time"
)

type ResourceStruct struct {
	ID          int64
	Name        string
	Text        string `prago-type:"text"`
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

type ResourceStructUnique struct {
	ID         int64
	UniqueName string `prago-unique:"true"`
}

func prepareResource() *Resource[ResourceStruct] {
	app := newTestingApp()
	resource := NewResource[ResourceStruct](app)
	NewResource[ResourceStructUnique](app)

	app.afterInit()
	app.unsafeDropTables()
	app.migrate(false)
	return resource
}

func prepareFuzzing() *Resource[ResourceStruct] {
	app := newTestingApp()
	resource := NewResource[ResourceStruct](app)
	NewResource[ResourceStructUnique](app)

	app.afterInit()
	//app.unsafeDropTables()
	//app.migrate(false)
	return resource
}

func TestBasicResource2(t *testing.T) {
	resource := prepareResource()

	item := &ResourceStruct{Name: "A", Floating: 3.14}

	err := resource.Create(item)
	if err != nil {
		t.Fatal(err)
	}

	if item.ID <= 0 {
		t.Fatal("should be positive")
	}

	resource.Create(&ResourceStruct{Name: "C"})
	resource.Create(&ResourceStruct{Name: "B"})

	list := resource.Query().List()
	if len(list) != 3 {
		t.Fatalf("wrong length %d", len(list))
	}

	first := resource.Query().Is("id", item.ID).First()
	if first.Name != "A" {
		t.Fatal("wrong name")
	}

	if resource.Is("id", item.ID).First().Name != "A" {
		t.Fatal("wrong name")
	}

	item.Name = "changed"

	err = resource.Update(item)
	if err != nil {
		t.Fatal(err)
	}

	if resource.Is("id", item.ID).First().Name != "changed" {
		t.Fatal("wrong name")
	}

	first = resource.Query().Is("name", "B").First()
	if first.Name != "B" {
		t.Fatal("wrong name")
	}

	count, _ := resource.Query().Count()
	if count != 3 {
		t.Fatalf("wrong count %d", count)
	}

	err = resource.Delete(item.ID)
	if err != nil {
		t.Fatal(err)
	}

	count, _ = resource.Query().Count()
	if count != 2 {
		t.Fatalf("wrong count %d", count)
	}

}

func TestQuery(t *testing.T) {
	resource := prepareResource()

	err := resource.Create(&ResourceStruct{Name: "A", Floating: 3.14})
	if err != nil {
		t.Fatal(err)
	}
	resource.Create(&ResourceStruct{Name: "C"})
	resource.Create(&ResourceStruct{Name: "B"})

	item := resource.Query().Where("id = ?", 2).First()
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	createdItem := resource.Query().Where("id = ?", 2).First()
	if createdItem == nil {
		t.Fatal("should not be nil")
	}
	if createdItem.Name != "C" {
		t.Fatal(createdItem.Name)
	}

	item = resource.Query().Where("id=?", 2).First()
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	item = resource.Query().First()
	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	if item.Floating < 3 || item.Floating > 4 {
		t.Fatal(item.Floating)
	}

	var list []*ResourceStruct
	list = resource.Query().List()
	if len(list) != 3 {
		t.Fatal(len(list))
	}

	if list[2].Name != "B" {
		t.Fatal(list[2].Name)
	}

	count, err := resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal(count)
	}

	list = resource.Query().Limit(1).Offset(1).Limit(1).List()
	if len(list) != 1 {
		t.Fatal(len(list))
	}
	if list[0].Name != "C" {
		t.Fatal(list[0].Name)
	}

	err = resource.Delete(item.ID)
	if err != nil {
		t.Fatal(err)
	}

	count, _ = resource.Query().Count()
	if count != 2 {
		t.Fatal(count)
	}
}

func TestResource(t *testing.T) {
	resource := prepareResource()
	items, err := resource.getListContent(&user{Role: "sysadmin"}, map[string][]string{
		"_order": {"id"},
	})
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

	err = resource.Create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	resource.Create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err = resource.Query().Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal(count)
	}

	items, err = resource.getListContent(&user{Role: "sysadmin"}, map[string][]string{
		"_order": {"id"},
		"_page":  {"1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(items.Rows[0].Items) != 4 {
		t.Fatalf("wrong length: %d", len(items.Rows[0].Items))
	}

	if items.Rows[1].Items[2].Value != "show" {
		t.Fatal(items.Rows[1].Items[2].Value)
	}
}

func TestResourceUnique(t *testing.T) {
	app := prepareResource().app

	resource := GetResource[ResourceStructUnique](app)

	resource.Create(&ResourceStructUnique{UniqueName: "A"})
	resource.Create(&ResourceStructUnique{UniqueName: "B"})
	resource.Create(&ResourceStructUnique{UniqueName: "A"})

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

	first := resource.Is("date", tm.Format("2006-01-02")).First()
	if first == nil {
		t.Fatal("should not be nil")
	}
}

func TestResourceTimestamps(t *testing.T) {
	resource := prepareResource()

	testStartTime := time.Now().Truncate(time.Second)

	resource.Create(&ResourceStruct{Name: "A"})

	item := resource.Query().Is("id", 1).First()

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

	trueItem := resource.Is("issomething", true).First()
	if trueItem.Name != "B" {
		t.Fatal(trueItem.Name)
	}

	falseItem := resource.Is("issomething", false).First()
	if falseItem.Name != "A" {
		t.Fatal(trueItem.Name)
	}
}

func TestResourceCreateWithID(t *testing.T) {
	resource := prepareResource()
	resource.Create(&ResourceStruct{ID: 85, Name: "A"})

	item := resource.Query().First()
	id := item.ID
	if id != 85 {
		t.Fatal(id)
	}
}

func TestShouldNotSaveWithZeroID(t *testing.T) {
	resource := prepareResource()
	err := resource.Update(&ResourceStruct{})
	if err == nil {
		t.Fatal("should not be nil")
	}
}

func TestWorkingWithConcreteID(t *testing.T) {
	resource := prepareResource()
	item := &ResourceStruct{
		ID:   3,
		Name: "A",
	}
	err := resource.Create(item)
	if err != nil {
		t.Fatal(err)
	}

	err = resource.Delete(3)
	if err != nil {
		t.Fatal(err)
	}

	item.Name = "B"

	err = resource.Create(item)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReplace(t *testing.T) {
	resource := prepareResource()
	var id int64 = 3
	item := &ResourceStruct{
		ID:   id,
		Name: "A",
	}
	err := resource.Replace(item)
	if err != nil {
		t.Fatal(err)
	}
	if resource.Is("id", id).First() == nil {
		t.Fatal("should not be nil")
	}
	item.Name = "B"
	err = resource.Replace(item)
	if err != nil {
		t.Fatal(err)
	}

	count, _ := resource.Query().Count()
	if count != 1 {
		t.Fatal(count)
	}

	modified := resource.Is("id", id).First()
	if modified.Name != "B" {
		t.Fatal(modified.Name)
	}
}

func FuzzCreateItem(f *testing.F) {
	f.Add(5, "helloss")
	resource := prepareFuzzing()
	f.Fuzz(func(t *testing.T, i int, s string) {
		item := &ResourceStruct{
			Name:  s,
			Count: int64(i),
		}
		err := resource.Create(item)
		if err != nil {
			return
		}
		item2 := resource.Query().Is("id", item.ID).First()
		if item2 == nil {
			t.Fatal("item2 is nil")
		}
		if item2.Name != item.Name {
			t.Fatal("name " + s)
		}
		if item2.Count != item.Count {
			t.Fatal("count ", i)
		}
	})
}

func TestLongSaveText(t *testing.T) {
	//TODO: make it work with 100000
	text := "some" + string(make([]byte, 10000))
	res := prepareResource()
	newItem := &ResourceStruct{Text: text}
	err := res.Create(newItem)
	if err != nil {
		t.Fatal(err)
	}
	item := res.Is("id", newItem.ID).First()

	if !strings.HasPrefix(item.Text, "some") {
		t.Fatal(item.Text)
	}
}
