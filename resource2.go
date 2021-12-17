package prago

import "fmt"

type Resource2[T any] struct{}

/*func (resource Resource2[T]) CreateItem() {
	item := reflect.New(T)
	fmt.Println(item)

}*/

func newResource2[T any]() Resource2[T] {
	return Resource2[T]{}
}

type XXX struct {
	A string
}

func resource2playground() {
	r := newResource2[XXX]()
	fmt.Println(r)

}
