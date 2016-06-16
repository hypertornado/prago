package admin

import (
	"database/sql"
	"errors"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/utils"
	"reflect"
)

var (
	ErrorDontHaveModel = errors.New("This resource does not have model")
)

type ResourceAction func(prago.Request, interface{}) bool

type DBProvider interface {
	DB() *sql.DB
}

type AdminResource struct {
	ID                 string
	Name               func(string) string
	Typ                reflect.Type
	ResourceController *prago.Controller
	Authenticate       Authenticatizer
	Pagination         int64
	OrderByColumn      string
	OrderDesc          bool
	HasModel           bool
	HasView            bool
	item               interface{}
	admin              DBProvider
	table              string
	StructCache        *StructCache
	AfterFormCreated   func(*Form, prago.Request, bool) *Form
	VisibilityFilter   StructFieldFilter
	EditabilityFilter  StructFieldFilter
	Actions            map[string]ActionBinder

	BeforeList   ResourceAction
	BeforeNew    ResourceAction
	BeforeCreate ResourceAction
	AfterCreate  ResourceAction
	BeforeDetail ResourceAction
	BeforeUpdate ResourceAction
	AfterUpdate  ResourceAction
	BeforeDelete ResourceAction
	AfterDelete  ResourceAction
}

func NewResource(item interface{}) (*AdminResource, error) {
	structCache, err := NewStructCache(item)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(item)
	defaultName := typ.Name()
	ret := &AdminResource{
		Name:              func(string) string { return defaultName },
		ID:                utils.ColumnName(defaultName),
		Typ:               typ,
		Authenticate:      AuthenticateAdmin,
		Pagination:        100000,
		HasModel:          true,
		HasView:           true,
		item:              item,
		StructCache:       structCache,
		VisibilityFilter:  DefaultVisibilityFilter,
		EditabilityFilter: DefaultEditabilityFilter,
	}

	ret.Actions = map[string]ActionBinder{
		"list":   BindList,
		"order":  BindOrder,
		"new":    BindNew,
		"create": BindCreate,
		"detail": BindDetail,
		"update": BindUpdate,
		"delete": BindDelete,
	}

	ret.OrderByColumn, ret.OrderDesc = structCache.GetDefaultOrder()

	ifaceName, ok := item.(interface {
		AdminName(string) string
	})
	if ok {
		ret.Name = ifaceName.AdminName
	}

	ifaceID, ok := item.(interface {
		AdminID() string
	})
	if ok {
		ret.ID = ifaceID.AdminID()
	}

	ifaceHasAuthenticate, ok := item.(interface {
		Authenticate(*User) bool
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

	ifaceAdminAfterFormCreated, ok := item.(interface {
		AdminAfterFormCreated(*Form, prago.Request, bool) *Form
	})
	if ok {
		ret.AfterFormCreated = ifaceAdminAfterFormCreated.AdminAfterFormCreated
	}

	return ret, nil
}

func (a *Admin) initResource(resource *AdminResource) error {

	resource.ResourceController = a.AdminController.SubController()

	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("admin_resource", resource)
		next()

		if !request.IsProcessed() && request.GetData("statusCode") == nil {
			prago.Render(request, 200, "admin_layout")
		}
	})

	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		user := request.GetData("currentuser").(*User)
		if !resource.Authenticate(user) {
			request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_403"))
			request.SetData("admin_yield", "admin_message")
			prago.Render(request, 403, "admin_layout")
		} else {
			next()
		}
	})

	init, ok := resource.item.(interface {
		AdminInitResource(*Admin, *AdminResource) error
	})

	if ok {
		err := init.AdminInitResource(a, resource)
		if err != nil {
			return err
		}
	}
	return AdminInitResourceDefault(a, resource)
}

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.DB()
}

func (ar *AdminResource) tableName() string {
	return ar.table
}

func (ar *AdminResource) UnsafeDropTable() error {
	return dropTable(ar.db(), ar.tableName())
}

func (ar *AdminResource) Migrate() error {
	_, err := getTableDescription(ar.db(), ar.tableName())
	if err == nil {
		return migrateTable(ar.db(), ar.tableName(), ar.StructCache)
	} else {
		return createTable(ar.db(), ar.tableName(), ar.StructCache)
	}
}
