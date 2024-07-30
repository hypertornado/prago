package prago

import (
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
			ID int64
			//Name string `prago-validations:"unique"`
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

		/*if _, valid := TestValidationUpdate(app, &TStruct{}, app.testUserData("")); valid != false {
			t.Fatal("expected")
		}

		if _, valid := TestValidationUpdate(app, &TStruct{
			Name: "AAA",
		}, app.testUserData("")); valid != true {
			t.Fatal("expected")
		}*/
	})

}
