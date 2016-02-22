package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"github.com/jinzhu/gorm"
	"net/url"
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

type AdminRowItem struct {
	Name      string
	NameHuman string
	Template  string
	Value     interface{}
}

func (ar *AdminResource) Get(id int64) (interface{}, []AdminRowItem, error) {
	var item interface{}
	getItem(ar.db(), ar.tableName(), ar.Typ, &item, id)
	items, err := ar.getItems(reflect.ValueOf(item))
	return item, items, err
}

func (ar *AdminResource) GetItems() ([]AdminRowItem, error) {
	return ar.getItems(reflect.New(ar.Typ).Elem())
}

func (ar *AdminResource) getItems(itemVal reflect.Value) ([]AdminRowItem, error) {
	items := []AdminRowItem{}

	for i := 0; i < ar.Typ.NumField(); i++ {
		field := ar.Typ.Field(i)
		structItem := AdminRowItem{
			Name:      field.Name,
			NameHuman: field.Name,
			Template:  "admin_item_input",
		}

		reflect.ValueOf(&structItem.Value).Elem().Set(itemVal.Field(i))

		if structItem.Name != "ID" {
			items = append(items, structItem)
		}
	}

	return items, nil
}

func (ar *AdminResource) CreateItemFromParams(params url.Values) error {
	var item interface{}
	val := reflect.New(ar.Typ)
	reflect.ValueOf(&item).Elem().Set(val)

	fmt.Println(item)
	fmt.Println(reflect.TypeOf(item))
	bindData(item, params)
	return createItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) UpdateItemFromParams(id int64, params url.Values) error {
	item, _, err := ar.Get(id)
	if err != nil {
		return err
	}

	fmt.Println(item)
	fmt.Println(reflect.TypeOf(item))
	bindData(&item, params)
	//return saveItem(ar.db(), ar.tableName(), item)
	return nil
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
