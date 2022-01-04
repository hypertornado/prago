package prago

import (
	"fmt"
	"reflect"
)

type App2 struct {
	resources []any
}

func AddResource2[T any](app *App2) {
	var r = Resources2[T]{}
	app.resources = append(app.resources, r)
}

func GetResource2[T any](app *App2) Resources2[T] {
	resource := app.resources[0]
	fmt.Println(reflect.TypeOf(resource).Name())
	return resource.(Resources2[T])
}

type Resources2[T any] struct {
}

type Orange struct {
	isOrange bool
}

type Apple struct {
}

func resource2playground() {
	if true {
		return
	}

	a2 := new(App2)
	AddResource2[Orange](a2)
	orange := GetResource2[Orange](a2)
	fmt.Println("---")
	//orange.isOrange = true
	fmt.Println(orange)
	fmt.Println("---")
}
