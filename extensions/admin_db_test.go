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

func TestAdminTime(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

	n0 := &TestNode{Changed: time.Now()}
	createItem(db, tableName, n0)

	n1 := TestNode{}
	getItem(db, tableName, reflect.TypeOf(TestNode{}), &n1, n0.ID)

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

	BindData(&in, values, BindDataFilterDefault)

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
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

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
	getFirstItem(db, tableName, reflect.TypeOf(TestNode{}), node, listQuery{})

	if node.Name != "A" {
		t.Fatal(node.Name)
	}

}

func TestAdminDB(t *testing.T) {
	var err error
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

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
	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})

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
	BindData(item, values, BindDataFilterDefault)
	createItem(db, tableName, item)
	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})
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

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodesIface, listQuery{})
	nodes = nodesIface.([]TestNode)

	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
}

/*func TestAdminPokus(t *testing.T) {
	//var str = "pokus"
	//var item interface{} = str

	var constructed interface{}

	newVal := reflect.New(reflect.TypeOf(""))
	reflect.ValueOf(&constructed).Elem().Set(newVal)

	fmt.Println(reflect.TypeOf(constructed))
	//fmt.Println(reflect.ValueOf(constructed).Elem().Elem().Type())
	fmt.Println(constructed)
}*/

func TestReflect(t *testing.T) {
	var i interface{} = createStruct()
	changeStruct(&i)
	if i.(*TestNode).Name != "CHANGED" {
		t.Fatal("not changed")
	}
}

func createStruct() interface{} {
	return &TestNode{Name: "No"}
}

func changeStruct(i interface{}) {
	val := reflect.ValueOf(i).Elem().Elem().Elem()
	val.FieldByName("Name").SetString("CHANGED")
}

func TestAdminResourceQuery(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

	q := &adminResourceQuery{
		query:         listQuery{},
		db:            db,
		tableName:     tableName,
		sliceItemType: reflect.TypeOf(TestNode{}),
	}

	n0 := &TestNode{Name: "A1", OK: false}

	err := q.Save(n0)
	if err != nil {
		t.Fatal(err)
	}

	firstItem, err := q.Where(map[string]interface{}{"id": 1}).First()
	if err != nil {
		t.Fatal(err)
	}

	var ii interface{}

	ii = &firstItem

	BindName(&ii)

	n, ok := firstItem.(*TestNode)
	if !ok {
		t.Fatal("not node pointer type - ", reflect.TypeOf(firstItem))
	}

	if n.Name != "A1" {
		t.Fatal(n.Name)
	}

}

func TestAdminDBList(t *testing.T) {
	tableName := "node"
	dropTable(db, tableName)
	createTable(db, tableName, reflect.TypeOf(TestNode{}))

	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now().Add(1 * time.Minute)})
	createItem(db, tableName, &TestNode{Name: "A"})
	createItem(db, tableName, &TestNode{Name: "B", Changed: time.Now()})
	createItem(db, tableName, &TestNode{Name: "C"})

	var nodes []*TestNode

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{})
	compareResults(t, nodes, []int64{1, 2, 3, 4})

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "id", asc: false}},
	})
	compareResults(t, nodes, []int64{4, 3, 2, 1})

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "name", asc: false}, {name: "changed", asc: true}},
	})
	compareResults(t, nodes, []int64{4, 3, 1, 2})

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		order: []listQueryOrder{{name: "name", asc: false}, {name: "changed", asc: false}},
	})
	compareResults(t, nodes, []int64{4, 1, 3, 2})

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		offset: 1,
		limit:  2,
	})
	compareResults(t, nodes, []int64{2, 3})

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
		whereString: "name=?",
		whereParams: []interface{}{"B"},
	})
	compareResults(t, nodes, []int64{1, 3})

	whereString, whereParams := mapToDBQuery(map[string]interface{}{"name": "B"})
	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{
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

	var node TestNode
	getFirstItem(db, tableName, reflect.TypeOf(TestNode{}), &node, listQuery{
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

	listItems(db, tableName, reflect.TypeOf(TestNode{}), &nodes, listQuery{})
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
