package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"reflect"
)

type AdminResource struct {
	ID    string
	Name  string
	Item  interface{}
	Items []*AdminResourceItem
	admin *Admin
}

func NewResource(item interface{}) (*AdminResource, error) {
	name := reflect.TypeOf(item).Name()

	ret := &AdminResource{
		Item: item,
		Name: name,
		ID:   utils.PrettyUrl(name),
	}
	return ret, nil
}

type mysqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   sql.NullString
}

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.db
}

func (ar *AdminResource) tableName() string {
	return ar.ID
}

type listResult struct {
	Url  string
	Name string
}

func (ar *AdminResource) List() ([]listResult, error) {
	ret := []listResult{}
	q := fmt.Sprintf("SELECT id, name FROM `%s`;", ar.tableName())

	rows, err := ar.db().Query(q)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		rows.Scan(&id, &name)
		item := &listResult{
			Url:  fmt.Sprintf("%s/%s/%d", ar.admin.Prefix, ar.ID, id),
			Name: name,
		}
		ret = append(ret, *item)
	}

	return ret, nil
}

func (ar *AdminResource) CreateItem(item interface{}) error {
	typ := reflect.TypeOf(item)
	id := utils.PrettyUrl(typ.Name())
	if id != ar.tableName() {
		return errors.New("Wrong class of item")
	}

	return createItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) Migrate() error {
	var err error
	fmt.Println("Migrating ", ar.Name, ar.ID)
	err = dropTable(ar.db(), ar.tableName())
	if err != nil {
		fmt.Println(err)
	}

	_, err = getTableDescription(ar.db(), ar.tableName())
	return createTable(ar.db(), ar.tableName(), ar.Item)
}
