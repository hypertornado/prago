package prago

import (
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

func prepareResource(initFns ...func(app *App)) (*App, *Resource) {
	var resource *Resource
	app := NewTestingApp(func(app *App) {
		resource = app.CreateResource(ResourceStruct{}, nil)
		for _, v := range initFns {
			v(app)
		}
	})
	app.unsafeDropTables()
	app.migrate(false)
	return app, resource
}

func TestQuery(t *testing.T) {
	var item ResourceStruct
	var createdItem interface{}
	var resource *Resource

	app, resource := prepareResource()

	err := app.Create(&ResourceStruct{Name: "A", Floating: 3.14})
	if err != nil {
		t.Fatal(err)
	}
	app.Create(&ResourceStruct{Name: "C"})
	app.Create(&ResourceStruct{Name: "B"})

	err = app.Query().Where(2).Get(&item)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	resource.newItem(&createdItem)
	err = app.Query().Where(2).Get(createdItem)
	if err != nil {
		t.Fatal(err)
	}
	if createdItem.(*ResourceStruct).Name != "C" {
		t.Fatal(createdItem.(*ResourceStruct).Name)
	}

	err = app.Query().Where("id=?", 2).Get(&item)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "C" {
		t.Fatal(item.Name)
	}

	app.Query().Get(&item)
	if item.Name != "A" {
		t.Fatal(item.Name)
	}

	if item.Floating < 3 || item.Floating > 4 {
		t.Fatal(item.Floating)
	}

	var list []*ResourceStruct
	err = app.Query().Get(&list)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 3 {
		t.Fatal(len(list))
	}

	if list[2].Name != "B" {
		t.Fatal(list[2].Name)
	}

	count, err := app.Query().Count(&ResourceStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal(count)
	}

	app.Query().Limit(1).Offset(1).Limit(1).Get(&list)
	if len(list) != 1 {
		t.Fatal(len(list))
	}
	if list[0].Name != "C" {
		t.Fatal(list[0].Name)
	}

	if count, _ = app.Query().WhereIs("name", "A").Delete(&ResourceStruct{}); count != 1 {
		t.Fatal(count)
	}

	if count, _ = app.Query().Count(&ResourceStruct{}); count != 2 {
		t.Fatal(count)
	}

}

func TestResource(t *testing.T) {
	app, resource := prepareResource()
	items, err := resource.getListContent(app, User{}, map[string][]string{
		"_order": {"id"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var item interface{}
	resource.newItem(&item)
	count, err := app.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal(count)
	}

	err = app.Create(&ResourceStruct{Name: "First", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	app.Create(&ResourceStruct{Name: "Second", Showing: "show"})

	count, err = app.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}

	items, _ = resource.getListContent(app, User{}, map[string][]string{
		"_order": {"id"},
		"_page":  {"1"},
	})

	if len(items.Rows[0].Items) != 2 {
		t.Fatal("wrong length")
	}

	if items.Rows[1].Items[1].Value != "show" {
		t.Fatal(items.Rows[1].Items[1].Value)
	}
}

func TestResourceUnique(t *testing.T) {
	type ResourceStructUnique struct {
		ID   int64
		Name string `prago-unique:"true"`
	}

	var resource *Resource
	app, _ := prepareResource(func(a *App) {
		resource = a.CreateResource(ResourceStructUnique{}, nil)
	})
	app.unsafeDropTables()
	app.migrate(false)

	app.Create(&ResourceStructUnique{Name: "A"})
	app.Create(&ResourceStructUnique{Name: "B"})
	app.Create(&ResourceStructUnique{Name: "A"})

	var item interface{}
	resource.newItem(&item)
	count, err := app.Query().Count(item)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatal(count)
	}
}

func TestResourceDate(t *testing.T) {
	app, resource := prepareResource()
	tm := time.Now()

	app.Create(&ResourceStruct{Date: tm})

	var item interface{}
	resource.newItem(&item)
	err := app.Query().WhereIs("date", tm.Format("2006-01-02")).Get(item)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResourceTimestamps(t *testing.T) {
	app, resource := prepareResource()

	testStartTime := time.Now().Truncate(time.Second)

	app.Create(&ResourceStruct{Name: "A"})

	var itemIface interface{}
	resource.newItem(&itemIface)
	err := app.Query().WhereIs("id", 1).Get(itemIface)
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
	app, resource := prepareResource()

	app.Create(&ResourceStruct{Name: "A", IsSomething: false})
	app.Create(&ResourceStruct{Name: "B", IsSomething: true})

	var itemIface interface{}
	resource.newItem(&itemIface)
	err := app.Query().WhereIs("issomething", true).Get(itemIface)

	if err != nil {
		t.Fatal(err)
	}

	item := itemIface.(*ResourceStruct)
	if item.Name != "B" {
		t.Fatal(item)
	}

	err = app.Query().WhereIs("issomething", false).Get(itemIface)
	if err != nil {
		t.Fatal(err)
	}

	item = itemIface.(*ResourceStruct)
	if item.Name != "A" {
		t.Fatal(item)
	}
}

func TestResourceCreateWithID(t *testing.T) {
	app, resource := prepareResource()
	app.Create(&ResourceStruct{ID: 85, Name: "A"})

	var item interface{}
	resource.newItem(&item)

	app.Query().Get(item)
	id := item.(*ResourceStruct).ID
	if id != 85 {
		t.Fatal(id)
	}
}

func TestShouldNotSaveWithZeroID(t *testing.T) {
	app, _ := prepareResource()
	err := app.Save(&ResourceStruct{})
	if err == nil {
		t.Fatal("should not be nil")
	}

}

/*
func TestLongSaveText(t *testing.T) {
	text := "some" + string(make([]byte, 100000))
	app, _ := prepareResource()
	err := app.Create(&ResourceStruct{Text: text})
	if err != nil {
		t.Fatal(err)
	}
	var item ResourceStruct
	app.Query().WhereIs("id", 1).Get(&item)

	if !strings.HasPrefix(item.Text, "some") {
		t.Fatal(item.Text)
	}
}*/
