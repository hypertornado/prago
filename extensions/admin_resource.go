package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
	"time"
)

type AdminResource struct {
	ID                 string
	Name               string
	Typ                reflect.Type
	ResourceController *prago.Controller
	item               interface{}
	admin              *Admin
}

func NewResource(item interface{}) (*AdminResource, error) {
	typ := reflect.TypeOf(item)
	name := typ.Name()
	ret := &AdminResource{
		Name: name,
		ID:   utils.PrettyUrl(name),
		Typ:  typ,
		item: item,
	}

	ifaceName, ok := item.(interface {
		AdminName() string
	})
	if ok {
		ret.Name = ifaceName.AdminName()
	}

	ifaceID, ok := item.(interface {
		AdminID() string
	})
	if ok {
		ret.ID = ifaceID.AdminID()
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

func (ar *AdminResource) ResourceURL(suffix string) string {
	ret := ar.admin.Prefix + "/" + ar.ID
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

type AdminFormItem struct {
	Name      string
	NameHuman string
	Template  string
	Error     string
	Value     interface{}
}

func (ar *AdminResource) GetFormItems(item interface{}) ([]AdminFormItem, error) {
	init, ok := ar.item.(interface {
		GetFormItems(*AdminResource, interface{}) ([]AdminFormItem, error)
	})

	if ok {
		return init.GetFormItems(ar, item)
	} else {
		return GetFormItemsDefault(ar, item)
	}
}

func GetFormItemsDefault(ar *AdminResource, item interface{}) ([]AdminFormItem, error) {
	itemVal := reflect.ValueOf(item)
	items := []AdminFormItem{}

	for i := 0; i < ar.Typ.NumField(); i++ {
		field := ar.Typ.Field(i)
		structItem := AdminFormItem{
			Name:      field.Name,
			NameHuman: field.Name,
			Template:  "admin_item_input",
		}

		reflect.ValueOf(&structItem.Value).Elem().Set(itemVal.Field(i))

		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Now()) {
				structItem.Template = "admin_item_date"
				var tm time.Time
				reflect.ValueOf(&tm).Elem().Set(reflect.ValueOf(structItem.Value))
				newVal := reflect.New(reflect.TypeOf("")).Elem()
				newVal.SetString(tm.Format("2006-01-02"))
				reflect.ValueOf(&structItem.Value).Elem().Set(newVal)
			}
		case reflect.Bool:
			structItem.Template = "admin_item_checkbox"
		case reflect.String:
			switch field.Tag.Get("prago-admin-type") {
			case "text":
				structItem.Template = "admin_item_textarea"
			}
		}

		description := field.Tag.Get("prago-admin-description")
		if len(description) > 0 {
			structItem.NameHuman = description
		}

		accessTag := field.Tag.Get("prago-admin-access")
		if accessTag == "-" || structItem.Name == "CreatedAt" || structItem.Name == "UpdatedAt" {
			structItem.Template = "admin_item_readonly"
		}

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
	bindData(item, params)
	return createItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) UpdateItemFromParams(id int64, params url.Values) error {

	var item interface{}
	val := reflect.New(ar.Typ)
	reflect.ValueOf(&item).Elem().Set(val)

	err := getItem(ar.db(), ar.tableName(), ar.Typ, item, id)
	if err != nil {
		return err
	}

	bindData(item, params)
	return saveItem(ar.db(), ar.tableName(), item)
}

func (ar *AdminResource) CreateItem(item interface{}) error {
	typ := reflect.TypeOf(item)
	id := utils.PrettyUrl(typ.Elem().Name())
	if id != ar.tableName() {
		return errors.New("Wrong class of item " + id + " " + ar.tableName())
	}

	return createItem(ar.db(), ar.tableName(), item)
}

//TODO: dont drop table
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
