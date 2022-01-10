package prago

import (
	"fmt"
	"reflect"
)

type Resource2[T any] struct {
	resource *Resource
}

func NewResource2[T any](app *App) *Resource2[T] {
	var item T
	app.Resource(item)
	ret := &Resource2[T]{}
	return ret
}

//func GetResource2[T any](app *App) *Resource2[T] {
//return nil
//}

// -----

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

func (r Resources2[T]) Create() *T {
	var ret T
	return &ret
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
	orangeResource := GetResource2[Orange](a2)
	orange := orangeResource.Create()
	fmt.Println("---")
	//orange.isOrange = true
	fmt.Println(orange.isOrange)
	fmt.Println("---")
}
