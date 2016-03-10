package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
	//"github.com/jinzhu/gorm"
	"reflect"
	"time"
)

var (
	ErrorDontHaveModel = errors.New("This resource does not have model")
)

type DBProvider interface {
	DB() *sql.DB
}

type AdminResource struct {
	ID                 string
	Name               string
	Typ                reflect.Type
	ResourceController *prago.Controller
	item               interface{}
	admin              DBProvider
	hasModel           bool
	hasView            bool
	queryFilter        func(*ResourceQuery) *ResourceQuery
	adminStructCache   *AdminStructCache
}

func NewResource(item interface{}) (*AdminResource, error) {
	structCache, err := NewAdminStructCache(item)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(item)
	name := typ.Name()
	ret := &AdminResource{
		Name:             name,
		ID:               utils.PrettyUrl(name),
		Typ:              typ,
		item:             item,
		hasModel:         true,
		hasView:          true,
		adminStructCache: structCache,
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

	ifaceHasModel, ok := item.(interface {
		AdminHasModel() bool
	})
	if ok {
		ret.hasModel = ifaceHasModel.AdminHasModel()
	}

	ifaceHasView, ok := item.(interface {
		AdminHasView() bool
	})
	if ok {
		ret.hasView = ifaceHasView.AdminHasView()
	}

	ret.queryFilter = QueryFilterDefault

	ifaceHasQueryFilter, ok := item.(interface {
		AdminQueryFilter(*ResourceQuery) *ResourceQuery
	})
	if ok {
		ret.queryFilter = ifaceHasQueryFilter.AdminQueryFilter
	}

	return ret, nil
}

/*func (ar *AdminResource) gorm() *gorm.DB {
	return ar.admin.gorm
}*/

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.DB()
}

func (ar *AdminResource) tableName() string {
	return ar.ID
}

func QueryFilterDefault(q *ResourceQuery) *ResourceQuery {
	q.Order("id")
	return q
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
		return ar.adminStructCache.GetFormItemsDefault(ar, item)
	}
}

type ItemCell struct {
	TemplateName string
	Value        interface{}
}

type ListTableRow struct {
	ID    int64
	Items []ItemCell
}

type ListTableHeader struct {
	Name      string
	NameHuman string
}

type ListTable struct {
	Header []ListTableHeader
	Rows   []ListTableRow
}

func (resource *AdminResource) ListTableItems() (table ListTable, err error) {
	q := resource.Query()
	q = resource.queryFilter(q)
	rowItems, err := q.List()

	for i := 0; i < resource.Typ.NumField(); i++ {
		field := resource.Typ.Field(i)
		tag := field.Tag.Get("prago-admin-show")
		if len(tag) > 0 || field.Name == "ID" || field.Name == "Name" {
			table.Header = append(table.Header, ListTableHeader{Name: field.Name, NameHuman: field.Name})
		}
	}

	val := reflect.ValueOf(rowItems)
	for i := 0; i < val.Len(); i++ {
		row := ListTableRow{}
		itemVal := val.Index(i).Elem()

		for _, h := range table.Header {
			structField, _ := resource.Typ.FieldByName(h.Name)
			fieldVal := itemVal.FieldByName(h.Name)
			row.Items = append(row.Items, ValueToCell(structField, fieldVal))
		}
		row.ID = itemVal.FieldByName("ID").Int()
		table.Rows = append(table.Rows, row)
	}
	return
}

func ValueToCell(field reflect.StructField, val reflect.Value) (cell ItemCell) {
	cell.TemplateName = "admin_string"
	var item interface{}
	reflect.ValueOf(&item).Elem().Set(val)
	cell.Value = item

	if field.Tag.Get("prago-admin-type") == "image" {
		cell.TemplateName = "admin_image"
	}

	if val.Type() == reflect.TypeOf(time.Now()) {
		var tm time.Time
		reflect.ValueOf(&tm).Elem().Set(val)
		cell.Value = tm.Format("2006-01-02 15:04:05")
	}

	return

	/*
		reflect.ValueOf(&item).Elem().Set(val)
		switch val.Kind() {
		case reflect.String:
			return fmt.Sprintf("%s", item)
		case reflect.Int64:
			return fmt.Sprintf("%d", item)
		case reflect.Bool:
			var b bool = item.(bool)
			if b {
				return "✔"
			} else {
				return "x"
			}
		}

		if val.Type() == reflect.TypeOf(time.Now()) {
			var tm time.Time
			reflect.ValueOf(&tm).Elem().Set(val)
			return tm.Format("2006-01-02 15:04:05")
		}

		return fmt.Sprintf("%s", item)
	*/
}

func (ar *AdminResource) Migrate() error {
	var err error
	//fmt.Println("Migrating ", ar.Name, ar.ID)
	err = dropTable(ar.db(), ar.tableName())
	if err != nil {
		fmt.Println(err)
	}

	_, err = getTableDescription(ar.db(), ar.tableName())
	return createTable(ar.db(), ar.tableName(), ar.adminStructCache)
}
