package extensions

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
	"testing"
	"time"
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
	Description string `prago-admin-type:"text"`
	OK          bool
	Count       int64
	Changed     time.Time
}

func TestAdminDB(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

	name1 := "A2"
	var count1 int64 = 13

	n0 := &TestNode{Name: "A1", OK: false}
	n1 := &TestNode{Name: name1, Count: count1, OK: true}

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

	n2 := TestNode{}

	err = getItem(db, tableName, reflect.TypeOf(TestNode{}), &n2, n1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if n2.Name != name1 {
		t.Fatal(n2.Name)
	}

	if n2.Count != count1 {
		t.Fatal(n2.Count)
	}

	if n2.OK != true {
		t.Fatal(n2.OK)
	}

	var nodesIface interface{}
	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface)

	nodes := nodesIface.([]TestNode)

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}

	if nodes[1].Name != name1 {
		t.Fatal(nodes[1].Name)
	}

	if nodes[1].Count != count1 {
		t.Fatal(nodes[1].Count)
	}

	var item interface{}
	val := reflect.New(reflect.TypeOf(TestNode{}))
	reflect.ValueOf(&item).Elem().Set(val)

	values := make(url.Values)
	values.Set("Name", "somename")
	bindData(item, values)
	createItem(db, tableName, item)
	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface)
	nodes = nodesIface.([]TestNode)

	if len(nodes) != 3 {
		t.Fatal(len(nodes))
	}

	changedNode := &TestNode{ID: 2, Name: "changedname"}
	saveItem(db, tableName, changedNode)

	var changedNodeResult TestNode
	getItem(db, tableName, reflect.TypeOf(TestNode{}), &changedNodeResult, 2)

	if changedNodeResult.Name != "changedname" {
		t.Fatal(changedNodeResult.Name)
	}

	deleteItem(db, tableName, 2)

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface)
	nodes = nodesIface.([]TestNode)

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
}

func TestAdminStructDescription(t *testing.T) {
	getStructDescription(reflect.TypeOf(&TestNode{}))

}

func f(i interface{}) {
	//typ := reflect.ValueOf(i).Elem().Type()
	reflect.ValueOf(i).Elem().FieldByName("ID").SetInt(54)
}

func TestAdminReflect(t *testing.T) {

	var item interface{}
	item = &TestNode{Name: "NAAME"}

	f(item)
}
