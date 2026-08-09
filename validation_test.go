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
			app.NewResource[TStruct]()
		})

		if _, valid := app.CheckUpdateValidation(&TStruct{}, app.TestUserData("")); valid != false {
			t.Fatal("expected")
		}

		if _, valid := app.CheckUpdateValidation(&TStruct{
			Name: "AAA",
		}, app.TestUserData("")); valid != true {
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
			app.NewResource[AStruct]()
			app.NewResource[RelatedStruct]()
		})

		var a = &AStruct{}

		must(app.Create(a))

		if _, ok := app.CheckDeleteValidation(a, app.TestUserData("")); !ok {
			t.Fatal("should be able to delete")
		}

		must(app.Create(&RelatedStruct{
			AStruct: a.ID,
		}))

		if _, ok := app.CheckDeleteValidation(a, app.TestUserData("")); ok {
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
			app.NewResource[AStruct]()
			app.NewResource[RelatedStruct]()
		})

		var a = &AStruct{}

		must(app.Create(a))

		if _, ok := app.CheckDeleteValidation(a, app.TestUserData("")); !ok {
			t.Fatal("should be able to delete")
		}

		must(app.Create(&RelatedStruct{
			AStruct: fmt.Sprintf(";%d;", a.ID),
		}))

		if _, ok := app.CheckDeleteValidation(a, app.TestUserData("")); ok {
			t.Fatal("should not be able to delete, because of strong relation")
		}
	})

	t.Run("enum field type validation", func(t *testing.T) {
		type AStruct struct {
			ID  int64
			Typ string `prago-type:"myenum"`
		}

		app := NewTesting(t, func(app *App) {
			app.AddEnumShort("myenum", [][2]string{
				{"a", "aname"},
				{"b", "bname"},
			})
			app.NewResource[AStruct]()
		})

		if _, valid := app.CheckUpdateValidation(&AStruct{Typ: ""}, app.TestUserData("")); valid != false {
			t.Fatal("should not be allowed")
		}
		if _, valid := app.CheckUpdateValidation(&AStruct{Typ: "c"}, app.TestUserData("")); valid != false {
			t.Fatal("should not be allowed")
		}
		if _, valid := app.CheckUpdateValidation(&AStruct{Typ: "a"}, app.TestUserData("")); valid == false {
			t.Fatal("should be ok")
		}

	})

}
