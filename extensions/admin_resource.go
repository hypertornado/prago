package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"github.com/jinzhu/gorm"
	"reflect"
)

type AdminResource struct {
	ID    string
	Name  string
	Typ   reflect.Type
	admin *Admin
}

func NewResource(item interface{}) (*AdminResource, error) {
	typ := reflect.TypeOf(item)
	name := typ.Name()

	ret := &AdminResource{
		Name: name,
		ID:   utils.PrettyUrl(name),
		Typ:  typ,
	}
	return ret, nil
}

func (ar *AdminResource) gorm() *gorm.DB {
	return ar.admin.gorm
}

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.db
}

func (ar *AdminResource) tableName() string {
	return ar.ID
}

func (ar *AdminResource) List() (interface{}, error) {
	var items interface{}
	listItems(ar.db(), ar.tableName(), ar.Typ, &items)
	return items, nil
}

func (ar *AdminResource) Get(id int64) (interface{}, error) {
	var item interface{}
	getItem(ar.db(), ar.tableName(), ar.Typ, &item, id)
	return item, nil
}

func (ar *AdminResource) CreateItem(item interface{}) error {
	typ := reflect.TypeOf(item)
	id := utils.PrettyUrl(typ.Elem().Name())
	if id != ar.tableName() {
		return errors.New("Wrong class of item " + id + " " + ar.tableName())
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
	return createTable(ar.db(), ar.tableName(), ar.Typ)
}
