package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"reflect"
	"time"
)

//ErrDontHaveModel is returned when item does not have a model
var ErrDontHaveModel = errors.New("resource does not have model")

/*type dbProvider interface {
	getDB() *sql.DB
	getResourceByName(string) *Resource
}*/

//Resource is structure representing one item in admin menu or one table in database
type Resource struct {
	Admin               *Admin
	ID                  string
	Name                func(locale string) string
	Typ                 reflect.Type
	ResourceController  *prago.Controller
	Authenticate        Authenticatizer
	Pagination          int64
	OrderByColumn       string
	OrderDesc           bool
	HasModel            bool
	HasView             bool
	item                interface{}
	table               string
	StructCache         *structCache
	AfterFormCreated    func(f *Form, request prago.Request, newItem bool) *Form
	VisibilityFilter    structFieldFilter
	EditabilityFilter   structFieldFilter
	resourceActions     []Action
	resourceItemActions []Action
	CanCreate           bool //TODO: should be based on user restrictions
	CanEdit             bool
	CanExport           bool

	ActivityLog bool

	PreviewURLFunction func(interface{}) string
}

//CreateResource creates new resource based on item
func (a *Admin) CreateResource(item interface{}, initFunction func(*Resource)) {
	cache, err := newStructCache(item, a.fieldTypes)
	if err != nil {
		panic(err)
	}

	typ := reflect.TypeOf(item)
	defaultName := typ.Name()
	ret := &Resource{
		Admin:              a,
		Name:               func(string) string { return defaultName },
		ID:                 columnName(defaultName),
		Typ:                typ,
		ResourceController: a.AdminController.SubController(),
		Authenticate:       AuthenticateAdmin,
		Pagination:         1000,
		HasModel:           true,
		HasView:            true,
		item:               item,
		StructCache:        cache,
		VisibilityFilter:   defaultVisibilityFilter,
		EditabilityFilter:  defaultEditabilityFilter,
		CanCreate:          true,
		CanEdit:            true,
		CanExport:          true,

		ActivityLog: true,
	}

	ret.OrderByColumn, ret.OrderDesc = cache.GetDefaultOrder()

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

	a.Resources = append(a.Resources, ret)
	if ret.HasModel {
		_, typFound := a.resourceMap[ret.Typ]
		if typFound {
			panic(fmt.Errorf("resource with type %s already created", ret.Typ))
		}

		a.resourceMap[ret.Typ] = ret
		a.resourceNameMap[ret.ID] = ret
	}

	if initFunction != nil {
		initFunction(ret)
	}
}

func (a *Admin) initResource(resource *Resource) {
	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("admin_resource", resource)
		next()
	})

	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		user := request.GetData("currentuser").(*User)
		if !resource.Authenticate(user) {
			render403(request)
		} else {
			next()
		}
	})

	_, ok := resource.item.(interface {
		InitResource(*Admin, *Resource) error
	})
	if ok {
		panic("use of deprecated API InitResource, " + resource.table)
		/*err := init.InitResource(a, resource)
		if err != nil {
			return err
		}*/
	}

	initResourceActions(a, resource)
}

func (a *Admin) getResourceByItem(item interface{}) (*Resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := a.resourceMap[typ]
	if !ok {
		return nil, fmt.Errorf("Can't find resource with type %s.", typ)
	}
	return resource, nil
}

func (ar *Resource) db() *sql.DB {
	return ar.Admin.getDB()
}

func (ar *Resource) tableName() string {
	return ar.table
}

func (ar *Resource) unsafeDropTable() error {
	return dropTable(ar.db(), ar.tableName())
}

func (ar *Resource) migrate(verbose bool) error {
	_, err := getTableDescription(ar.db(), ar.tableName())
	if err == nil {
		return migrateTable(ar.db(), ar.tableName(), ar.StructCache, verbose)
	}
	return createTable(ar.db(), ar.tableName(), ar.StructCache, verbose)
}

func (ar *Resource) saveWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() && val.FieldByName(fn).CanSet() && val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}

	return ar.StructCache.saveItem(db, ar.tableName(), item)
}

func (ar *Resource) createWithDBIface(item interface{}, db dbIface) error {
	if !ar.HasModel {
		return ErrDontHaveModel
	}

	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	var t time.Time
	for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
			reflect.ValueOf(&t).Elem().Set(field)
			if t.IsZero() {
				field.Set(timeVal)
			}
		}
	}
	return ar.StructCache.createItem(db, ar.tableName(), item)
}

func (ar *Resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(ar.Typ))
}

func (ar *Resource) newItems(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(reflect.SliceOf(reflect.PtrTo(ar.Typ))))
}
