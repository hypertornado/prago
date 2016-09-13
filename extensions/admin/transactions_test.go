package admin

import (
	"testing"
)

type TransactionTestStruct struct {
	ID    int64
	Name  string
	Count int64
}

func prepareTransactionResource() (*Admin, *Resource) {
	admin := prepareAdmin()
	resource, _ := admin.CreateResource(TransactionTestStruct{})
	admin.UnsafeDropTables()
	admin.Migrate(false)
	return admin, resource
}

func TestTransactions(t *testing.T) {
	admin, resource := prepareTransactionResource()

	s1 := TransactionTestStruct{Name: "a"}
	s2 := TransactionTestStruct{Name: "b"}

	t1 := admin.Transaction()

	var err error

	if err = t1.Create(&s1); err != nil {
		t.Fatal(err)
	}

	var c int64

	c, _ = resource.Query().Count()
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

	c, _ = resource.Query().Count()
	if c != 1 {
		t.Fatal(c)
	}

	admin.Create(&s1)
	admin.Create(&s2)

}
