package prago

import "testing"

func TestCache(t *testing.T) {
	resource := prepareResource()

	app := resource.data.app

	a := <-Cached(app, "xxx", func() string {
		return "A"
	})
	if a != "A" {
		t.Fatal(a)
	}

}
