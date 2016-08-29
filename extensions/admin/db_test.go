package admin

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

var (
	db          *sql.DB
	structCache *StructCache
)

type dbProvider struct{}

func (dbProvider) DB() *sql.DB {
	return db
}

func (dbProvider) GetResourceByName(string) *AdminResource {
	return nil
}

func init() {
	var err error
	structCache, err = NewStructCache(TestNode{})
	if err != nil {
		panic(err)
	}

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
	Description string `prago-type:"text"`
	OK          bool
	Count       int64
	Changed     time.Time
	Date        time.Time `prago-type:"date"`
}

func TestReflect(t *testing.T) {
	var i interface{}
	createStruct(&i)
	changeStruct(i)
	if i.(*TestStruct).Name != "Bar" {
		t.Fatal("not changed")
	}

	i = createStructFactory()
	changeStruct(i)
	if i.(*TestStruct).Name != "Bar" {
		t.Fatal("not changed")
	}
}

type TestStruct struct {
	Name string
}

func createStruct(i interface{}) {
	el := &TestStruct{Name: "Foo"}
	reflect.ValueOf(i).Elem().Set(reflect.ValueOf(el))
}

func createStructFactory() interface{} {
	var ret interface{}
	createStruct(&ret)
	return ret
}

func changeStruct(i interface{}) {
	val := reflect.ValueOf(&i).Elem().Elem().Elem()
	field := val.FieldByName("Name")
	field.SetString("Bar")
}

func TestAdminTime(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)

	createTable(db, tableName, structCache, false)

	tn := time.Now()

	n0 := &TestNode{Changed: tn, Date: tn}
	createItem(db, tableName, n0)

	n1 := &TestNode{}

	whereString, whereParams := mapToDBQuery(map[string]interface{}{"id": n0.ID})
	getFirstItem(structCache, db, tableName, &n1, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})

	if n1.Date.Hour() != 0 {
		t.Fatal(n1.Date.Hour())
	}

	if n1.Date.Minute() != 0 {
		t.Fatal(n1.Date.Minute())
	}

	if n1.Date.Second() != 0 {
		t.Fatal(n1.Date.Second())
	}

	if n0.Changed.Format("2006-01-02 15:04:05") != n1.Changed.Format("2006-01-02 15:04:05") {
		t.Fatal(n0.Changed, n1.Changed)
	}
}

func TestAdminBind(t *testing.T) {
	n := TestNode{}
	values := make(url.Values)
	values.Set("Name", "ABC")
	values.Set("Changed", "2014-11-10")

	var in interface{}
	in = &n

	structCache.BindData(&in, values, nil, DefaultEditabilityFilter)

	if n.Name != "ABC" {
		t.Fatal(n.Name)
	}
	if n.Changed.Format("2006-01-02") != "2014-11-10" {
		t.Fatal(n.Changed)
	}
}

func TestAdminDBFirst(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache, false)

	var node *TestNode
	err = getFirstItem(structCache, db, tableName, &node, &listQuery{})
	if err != ErrorNotFound {
		t.Fatal("wrong error")
	}

	n0 := &TestNode{Name: "A"}
	n1 := &TestNode{Name: "B"}

	err = createItem(db, tableName, n0)
	if err != nil {
		t.Fatal(err)
	}

	err = createItem(db, tableName, n1)
	if err != nil {
		t.Fatal(err)
	}

	getFirstItem(structCache, db, tableName, &node, &listQuery{})

	if node.Name != "A" {
		t.Fatal(node.Name)
	}
}

func TestAdminListItems(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache, false)

	createItem(db, tableName, &TestNode{Name: "A"})
	createItem(db, tableName, &TestNode{Name: "B"})

	var nodesIface interface{}
	listItems(structCache, db, tableName, &nodesIface, &listQuery{})

	nodes, ok := nodesIface.([]*TestNode)
	if !ok {
		t.Fatal(reflect.TypeOf(nodesIface))
	}

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}

	if (nodes)[0].Name != "A" {
		t.Fatal((nodes)[0].Name)
	}

	var nodeIface interface{}
	getFirstItem(structCache, db, tableName, &nodeIface, &listQuery{})

	_, ok = nodeIface.(*TestNode)
	if !ok {
		t.Fatal(reflect.TypeOf(nodeIface))
	}

	bindName(&nodeIface)

	if nodeIface.(*TestNode).Name != "CHANGED" {
		t.Fatal("Wrong changed")
	}
}

func bindName(item interface{}) {
	value := reflect.ValueOf(item).Elem().Elem().Elem()
	field := value.FieldByName("Name")
	field.SetString("CHANGED")
}

func TestAdminDB(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache, false)

	timeNow := time.Now()

	name1 := "A2"
	var count1 int64 = 13

	n0 := &TestNode{Name: "A1", OK: false, Changed: timeNow}
	n1 := &TestNode{Name: name1, Count: count1, OK: true, Changed: timeNow}

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

	whereString, whereParams := mapToDBQuery(map[string]interface{}{"id": n1.ID})
	err = getFirstItem(structCache, db, tableName, &n2, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
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
	listItems(structCache, db, tableName, &nodesIface, &listQuery{})

	nodes := nodesIface.([]*TestNode)

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
	structCache.BindData(item, values, nil, DefaultEditabilityFilter)
	createItem(db, tableName, item)
	listItems(structCache, db, tableName, &nodesIface, &listQuery{})
	nodes = nodesIface.([]*TestNode)

	if len(nodes) != 3 {
		t.Fatal(len(nodes))
	}

	changedNode := &TestNode{ID: 2, Name: "changedname"}
	saveItem(db, tableName, changedNode)

	changedNodeResult := &TestNode{}

	whereString, whereParams = mapToDBQuery(map[string]interface{}{"id": 2})
	getFirstItem(structCache, db, tableName, &changedNodeResult, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})

	if changedNodeResult.Name != "changedname" {
		t.Fatal(changedNodeResult.Name)
	}

	deleteItems(db, tableName, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})

	listItems(structCache, db, tableName, &nodesIface, &listQuery{})
	nodes = nodesIface.([]*TestNode)

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
}

func TestAdminResourceQuery(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache, false)

	q := &ResourceQuery{
		query:       &listQuery{},
		db:          db,
		tableName:   tableName,
		structCache: structCache,
	}

	n0 := &TestNode{Name: "A1", OK: false}

	createItem(db, tableName, n0)

	firstItem, err := q.Where(map[string]interface{}{"id": 1}).First()
	if err != nil {
		t.Fatal(err)
	}

	n, ok := firstItem.(*TestNode)
	if !ok {
		t.Fatal("bad type")
	}

	if n.Name != "A1" {
		t.Fatal(n.Name)
	}
}

func TestAdminDBList(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache, false)

	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now().Add(1 * time.Minute)})
	createItem(db, tableName, &TestNode{Name: "A"})
	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now()})
	createItem(db, tableName, &TestNode{Name: "C"})

	var nodes []*TestNode

	listItems(structCache, db, tableName, &nodes, &listQuery{})
	compareResults(t, nodes, []int64{1, 2, 3, 4})

	listItems(structCache, db, tableName, &nodes, &listQuery{
		order: []listQueryOrder{{name: "id", desc: true}},
	})
	compareResults(t, nodes, []int64{4, 3, 2, 1})

	listItems(structCache, db, tableName, &nodes, &listQuery{
		order: []listQueryOrder{{name: "name", desc: true}, {name: "changed", desc: false}},
	})
	compareResults(t, nodes, []int64{4, 3, 1, 2})

	listItems(structCache, db, tableName, &nodes, &listQuery{
		order: []listQueryOrder{{name: "name", desc: true}, {name: "changed", desc: true}},
	})
	compareResults(t, nodes, []int64{4, 1, 3, 2})

	listItems(structCache, db, tableName, &nodes, &listQuery{
		offset: 1,
		limit:  2,
	})
	compareResults(t, nodes, []int64{2, 3})

	listItems(structCache, db, tableName, &nodes, &listQuery{
		whereString: "name=?",
		whereParams: []interface{}{"B"},
	})
	compareResults(t, nodes, []int64{1, 3})

	whereString, whereParams := mapToDBQuery(map[string]interface{}{"name": "B"})
	listItems(structCache, db, tableName, &nodes, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
	compareResults(t, nodes, []int64{1, 3})

	var i int64

	i, _ = countItems(db, tableName, &listQuery{})
	if i != 4 {
		t.Fatal(i)
	}

	i, _ = countItems(db, tableName, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
	if i != 2 {
		t.Fatal(i)
	}

	var node *TestNode
	getFirstItem(structCache, db, tableName, &node, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
		offset:      1,
	})

	if node.ID != 3 {
		t.Fatal(node.ID)
	}

	whereString, whereParams = mapToDBQuery(map[string]interface{}{"name": "B", "id": 1})
	listItems(structCache, db, tableName, &nodes, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
	compareResults(t, nodes, []int64{1})

	whereString, whereParams = mapToDBQuery(map[string]interface{}{"name": "B"})
	count, err := deleteItems(db, tableName, &listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})

	if count != 2 {
		t.Fatal(count)
	}
	if err != nil {
		t.Fatal(err)
	}

	listItems(structCache, db, tableName, &nodes, &listQuery{})
	compareResults(t, nodes, []int64{2, 4})

}

func compareResults(t *testing.T, nodes []*TestNode, ids []int64) {
	if len(nodes) != len(ids) {
		t.Fatal("not equal length ", len(nodes))
	}

	for i, _ := range nodes {
		if nodes[i].ID != ids[i] {
			t.Fatal(nodes[i], ids)
		}
	}
}

type N1 struct {
	ID   int64
	Name string
}

type N2 struct {
	ID          int64
	Name        string
	Description string
}

func TestMigrateTable(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)

	structCache1, _ := NewStructCache(N1{})
	structCache2, _ := NewStructCache(N2{})

	createTable(db, tableName, structCache1, false)

	createItem(db, tableName, &N1{Name: "A"})

	err = migrateTable(db, tableName, structCache2, false)
	if err != nil {
		t.Fatal(err)
	}

	createItem(db, tableName, &N2{Name: "B", Description: "D"})

	var nodes []*N2
	err = listItems(structCache2, db, tableName, &nodes, &listQuery{})
	if err != nil {
		t.Fatal(err)
	}

	if nodes[0].Name != "A" {
		t.Fatal(nodes[0])
	}

	if nodes[1].Name != "B" {
		t.Fatal(nodes[1])
	}

	if nodes[1].Description != "D" {
		t.Fatal(nodes[1])
	}
}
