package prago

import (
	"testing"
)

func TestTransactions(t *testing.T) {
	resource := prepareResource2()
	app := resource.Resource.app

	s1 := ResourceStruct2{Name: "a"}
	s2 := ResourceStruct2{Name: "b"}

	t1 := app.Transaction()

	var err error

	if err = t1.Create(&s1); err != nil {
		t.Fatal(err)
	}

	var c int64

	c, _ = resource.Count()
	if c != 0 {
		t.Fatal(c)
	}

	c, _ = t1.Query().Count(&s1)
	if c != 1 {
		t.Fatal(c)
	}

	if err = t1.Commit(); err != nil {
		t.Fatal(err)
	}

	c, _ = resource.Count()
	if c != 1 {
		t.Fatal(c)
	}

	resource.Create(&s1)
	resource.Create(&s2)

}
