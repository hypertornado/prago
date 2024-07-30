package prago

import (
	"fmt"
	"testing"
)

func TestValidation(t *testing.T) {

	t.Run("nonempty validations", func(t *testing.T) {
		type TStruct struct {
			ID   int64
			Name string `prago-validations:"nonempty"`
		}

		app := NewTesting(t, func(app *App) {
			NewResource[TStruct](app)
		})

		if _, valid := TestValidationUpdate(app, &TStruct{}, app.testUserData("")); valid != false {
			t.Fatal("expected")
		}

		if _, valid := TestValidationUpdate(app, &TStruct{
			Name: "AAA",
		}, app.testUserData("")); valid != true {
			t.Fatal("expected")
		}
	})

	t.Run("strong validations", func(t *testing.T) {
		type AStruct struct {
			ID int64
		}

		type RelatedStruct struct {
			ID      int64
			AStruct int64 `prago-type:"relation" prago-validations:"strong"`
		}

		app := NewTesting(t, func(app *App) {
			NewResource[AStruct](app)
			NewResource[RelatedStruct](app)
		})

		var a = &AStruct{}

		must(CreateItem(app, a))

		if _, ok := TestValidationDelete(app, a, app.testUserData("")); !ok {
			t.Fatal("should be able to delete")
		}

		must(CreateItem(app, &RelatedStruct{
			AStruct: a.ID,
		}))

		if _, ok := TestValidationDelete(app, a, app.testUserData("")); ok {
			t.Fatal("should be able to delete, because of strong relation")
		}
	})

	t.Run("strong multirelation validations", func(t *testing.T) {
		type AStruct struct {
			ID int64
		}

		type RelatedStruct struct {
			ID      int64
			AStruct string `prago-type:"multirelation" prago-validations:"strong"`
		}

		app := NewTesting(t, func(app *App) {
			NewResource[AStruct](app)
			NewResource[RelatedStruct](app)
		})

		var a = &AStruct{}

		must(CreateItem(app, a))

		if _, ok := TestValidationDelete(app, a, app.testUserData("")); !ok {
			t.Fatal("should be able to delete")
		}

		must(CreateItem(app, &RelatedStruct{
			AStruct: fmt.Sprintf(";%d;", a.ID),
		}))

		if _, ok := TestValidationDelete(app, a, app.testUserData("")); ok {
			t.Fatal("should not be able to delete, because of strong relation")
		}
	})

	t.Run("enum field type validation", func(t *testing.T) {
		type AStruct struct {
			ID  int64
			Typ string `prago-type:"myenum"`
		}

		app := NewTesting(t, func(app *App) {
			app.AddEnumFieldType("myenum", [][2]string{
				{"a", "aname"},
				{"b", "bname"},
			})
			NewResource[AStruct](app)
		})

		if _, valid := TestValidationUpdate(app, &AStruct{Typ: ""}, app.testUserData("")); valid != false {
			t.Fatal("should not be allowed")
		}
		if _, valid := TestValidationUpdate(app, &AStruct{Typ: "c"}, app.testUserData("")); valid != false {
			t.Fatal("should not be allowed")
		}
		if _, valid := TestValidationUpdate(app, &AStruct{Typ: "a"}, app.testUserData("")); valid == false {
			t.Fatal("should be ok")
		}

	})

}
