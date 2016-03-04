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

var (
	db          *sql.DB
	structCache *AdminStructCache
)

type dbProvider struct{}

func (dbProvider) DB() *sql.DB {
	return db
}

func init() {
	var err error
	structCache, err = NewAdminStructCache(TestNode{})
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
	Description string `prago-admin-type:"text"`
	OK          bool
	Count       int64
	Changed     time.Time
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

	createTable(db, tableName, structCache)

	n0 := &TestNode{Changed: time.Now()}
	createItem(db, tableName, n0)

	n1 := &TestNode{}
	getItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &n1, n0.ID)

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

	BindData(&in, values, nil, BindDataFilterDefault)

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
	createTable(db, tableName, structCache)

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

	var node *TestNode = &TestNode{Name: "OLD"}
	getFirstItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &node, listQuery{})

	if node.Name != "A" {
		t.Fatal(node.Name)
	}
}

func TestAdminListItems(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache)

	createItem(db, tableName, &TestNode{Name: "A"})
	createItem(db, tableName, &TestNode{Name: "B"})

	var nodesIface interface{}
	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})

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
	getFirstItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodeIface, listQuery{})

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
	createTable(db, tableName, structCache)

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

	err = getItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &n2, n1.ID)
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
	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})

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
	BindData(item, values, nil, BindDataFilterDefault)
	createItem(db, tableName, item)
	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})
	nodes = nodesIface.([]*TestNode)

	if len(nodes) != 3 {
		t.Fatal(len(nodes))
	}

	changedNode := &TestNode{ID: 2, Name: "changedname"}
	saveItem(db, tableName, changedNode)

	changedNodeResult := &TestNode{}
	getItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &changedNodeResult, 2)

	if changedNodeResult.Name != "changedname" {
		t.Fatal(changedNodeResult.Name)
	}

	deleteItem(db, tableName, 2)

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})
	nodes = nodesIface.([]*TestNode)

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
}

func TestAdminResourceQuery(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, structCache)

	q := &ResourceQuery{
		query:       listQuery{},
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
	createTable(db, tableName, structCache)

	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now().Add(1 * time.Minute)})
	createItem(db, tableName, &TestNode{Name: "A"})
	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now()})
	createItem(db, tableName, &TestNode{Name: "C"})

	var nodes []*TestNode

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{})
	compareResults(t, nodes, []int64{1, 2, 3, 4})

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "id", desc: true}},
	})
	compareResults(t, nodes, []int64{4, 3, 2, 1})

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "name", desc: true}, {name: "changed", desc: false}},
	})
	compareResults(t, nodes, []int64{4, 3, 1, 2})

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "name", desc: true}, {name: "changed", desc: true}},
	})
	compareResults(t, nodes, []int64{4, 1, 3, 2})

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		offset: 1,
		limit:  2,
	})
	compareResults(t, nodes, []int64{2, 3})

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		whereString: "name=?",
		whereParams: []interface{}{"B"},
	})
	compareResults(t, nodes, []int64{1, 3})

	whereString, whereParams := mapToDBQuery(map[string]interface{}{"name": "B"})
	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
	compareResults(t, nodes, []int64{1, 3})

	var i int64

	i, _ = countItems(db, tableName, listQuery{})
	if i != 4 {
		t.Fatal(i)
	}

	i, _ = countItems(db, tableName, listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})
	if i != 2 {
		t.Fatal(i)
	}

	var node *TestNode
	getFirstItem(structCache, db, tableName, reflect.TypeOf(TestNode{}), &node, listQuery{
		whereString: whereString,
		whereParams: whereParams,
		offset:      1,
	})

	if node.ID != 3 {
		t.Fatal(node.ID)
	}

	whereString, whereParams = mapToDBQuery(map[string]interface{}{"name": "B"})
	count, err := deleteItems(db, tableName, listQuery{
		whereString: whereString,
		whereParams: whereParams,
	})

	if count != 2 {
		t.Fatal(count)
	}
	if err != nil {
		t.Fatal(err)
	}

	listItems(structCache, db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{})
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

func TestAdminReflect(t *testing.T) {
	/*tn := time.Now()
	fmt.Println(tn)
	var t2 time.Time
	reflect.ValueOf(&t2).Elem().Set(reflect.ValueOf(tn))
	fmt.Println(t2)*/
}
