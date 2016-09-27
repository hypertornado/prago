package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"reflect"
)

var ErrDontHaveModel = errors.New("resource does not have model")

type Action func(prago.Request, interface{}) bool

type dbProvider interface {
	getDB() *sql.DB
	getResourceByName(string) *Resource
}

type Resource struct {
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
	admin              dbProvider
	table              string
	StructCache        *StructCache
	AfterFormCreated   func(f *Form, request prago.Request, newItem bool) *Form
	VisibilityFilter   StructFieldFilter
	EditabilityFilter  StructFieldFilter
	Actions            map[string]ActionBinder

	BeforeList   Action
	BeforeNew    Action
	BeforeCreate Action
	AfterCreate  Action
	BeforeDetail Action
	BeforeUpdate Action
	AfterUpdate  Action
	BeforeDelete Action
	AfterDelete  Action
}

func NewResource(item interface{}) (*Resource, error) {
	structCache, err := NewStructCache(item)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(item)
	defaultName := typ.Name()
	ret := &Resource{
		Name:              func(string) string { return defaultName },
		ID:                columnName(defaultName),
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

func (a *Admin) initResource(resource *Resource) error {

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
			Render403(request)
		} else {
			next()
		}
	})

	init, ok := resource.item.(interface {
		InitResource(*Admin, *Resource) error
	})

	if ok {
		err := init.InitResource(a, resource)
		if err != nil {
			return err
		}
	}
	return InitResourceDefault(a, resource)
}

func (a *Admin) getResourceByItem(item interface{}) (*Resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := a.resourceMap[typ]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Can't find resource with type %s.", typ))
	}
	return resource, nil
}

func (ar *Resource) db() *sql.DB {
	return ar.admin.getDB()
}

func (ar *Resource) tableName() string {
	return ar.table
}

func (ar *Resource) UnsafeDropTable() error {
	return dropTable(ar.db(), ar.tableName())
}

func (ar *Resource) migrate(verbose bool) error {
	_, err := getTableDescription(ar.db(), ar.tableName())
	if err == nil {
		return migrateTable(ar.db(), ar.tableName(), ar.StructCache, verbose)
	} else {
		return createTable(ar.db(), ar.tableName(), ar.StructCache, verbose)
	}
}

func Render403(request prago.Request) {
	request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	prago.Render(request, 403, "admin_layout")
}

func Render404(request prago.Request) {
	request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	prago.Render(request, 404, "admin_layout")
}
