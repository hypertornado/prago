package prago

import (
	"testing"
)

func TestCache(t *testing.T) {
	resource := prepareResource(t)

	app := resource.app

	a := app.Cached("xxx", func() string {
		return "A"
	})
	if a != "A" {
		t.Fatal(a)
	}

}
