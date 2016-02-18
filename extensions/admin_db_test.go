package extensions

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"testing"
)

var db *sql.DB

func init() {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", "prago", "prago", "prago_test")
	g, err := gorm.Open("mysql", connectString)
	if err != nil {
		panic(err)
	}
	db = g.DB()
}

type TestNode struct {
	ID          int64
	Name        string
	Description string
	Count       int64
}

func TestAdminDB(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, TestNode{})

	name1 := "A2"
	var count1 int64 = 13

	n0 := &TestNode{Name: "A1"}
	n1 := &TestNode{Name: name1, Count: count1}

	err = createItem(db, tableName, n0)
	if err != nil {
		t.Fatal(err)
	}

	err = createItem(db, tableName, n1)
	if err != nil {
		t.Fatal(err)
	}

	if n1.ID != 2 {
		t.Fatal(n1.ID)
	}

	n2 := &TestNode{}

	err = getItem(db, tableName, n2, n1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if n2.Name != name1 {
		t.Fatal(n2.Name)
	}

	if n2.Count != count1 {
		t.Fatal(n2.Count)
	}

	var nodes []TestNode
	listItems(db, tableName, &nodes)

	fmt.Println(nodes)

}

func TestAdminStructDescription(t *testing.T) {
	getStructDescription(&TestNode{})
}
