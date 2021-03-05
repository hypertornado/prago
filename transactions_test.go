package prago

import (
	"testing"
)

func TestTransactions(t *testing.T) {
	admin, _ := prepareResource()

	s1 := ResourceStruct{Name: "a"}
	s2 := ResourceStruct{Name: "b"}

	t1 := admin.Transaction()

	var err error

	if err = t1.Create(&s1); err != nil {
		t.Fatal(err)
	}

	var c int64

	c, _ = admin.Query().Count(&ResourceStruct{})
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

	c, _ = admin.Query().Count(&ResourceStruct{})
	if c != 1 {
		t.Fatal(c)
	}

	admin.Create(&s1)
	admin.Create(&s2)

}
