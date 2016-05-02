package admin

import (
	"database/sql"
	"errors"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
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
	Authenticate       Authenticatizer
	item               interface{}
	admin              DBProvider
	hasModel           bool
	hasView            bool
	table              string
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
		Authenticate:     AuthenticateAdmin,
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

	ifaceHasAuthenticate, ok := item.(interface {
		Authenticate(User) bool
	})
	if ok {
		ret.Authenticate = ifaceHasAuthenticate.Authenticate
	}

	ifaceHasTableName, ok := item.(interface {
		AdminHasTableName() string
	})
	if ok {
		ret.table = ifaceHasTableName.AdminHasTableName()
	} else {
		ret.table = ret.ID
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

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.DB()
}

func (ar *AdminResource) tableName() string {
	return ar.table
}

func QueryFilterDefault(q *ResourceQuery) *ResourceQuery {
	q.Order("id")
	return q
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

func (resource *AdminResource) ListTableItems(lang string) (table ListTable, err error) {
	q := resource.Query()
	q = resource.queryFilter(q)
	rowItems, err := q.List()

	for _, v := range resource.adminStructCache.fieldArrays {
		showTag := v.tags["prago-preview"]
		if showTag == "true" || v.name == "ID" || v.name == "Name" {
			table.Header = append(table.Header, ListTableHeader{Name: v.name, NameHuman: v.humanName(lang)})
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

	if field.Tag.Get("prago-type") == "image" {
		cell.TemplateName = "admin_image"
	}

	if val.Type() == reflect.TypeOf(time.Now()) {
		var tm time.Time
		reflect.ValueOf(&tm).Elem().Set(val)
		cell.Value = tm.Format("2006-01-02 15:04:05")
	}
	return
}

func (ar *AdminResource) UnsafeDropTable() error {
	return dropTable(ar.db(), ar.tableName())
}

func (ar *AdminResource) Migrate() error {
	_, err := getTableDescription(ar.db(), ar.tableName())
	if err == nil {
		return migrateTable(ar.db(), ar.tableName(), ar.adminStructCache)
	} else {
		return createTable(ar.db(), ar.tableName(), ar.adminStructCache)
	}
}
