package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"reflect"
	"time"
)

//ErrDontHaveModel is returned when item does not have a model
var ErrDontHaveModel = errors.New("resource does not have model")

//Action represents functions which can be added before or after admin operations
type Action func(prago.Request, interface{}) bool

type dbProvider interface {
	getDB() *sql.DB
	getResourceByName(string) *Resource
}

type Snippet struct {
	Template string
}

//Resource is structure representing one item in admin menu or one table in database
type Resource struct {
	ID                 string
	Snippets           []Snippet
	Name               func(locale string) string
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
	StructCache        *structCache
	AfterFormCreated   func(f *Form, request prago.Request, newItem bool) *Form
	VisibilityFilter   structFieldFilter
	EditabilityFilter  structFieldFilter
	Actions            map[string]ActionBinder
	ResourceActions    []ResourceAction
	CanCreate          bool
	CanEdit            bool

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

//CreateResource creates new resource based on item
func (a *Admin) CreateResource(item interface{}) (ret *Resource, err error) {
	cache, err := newStructCache(item)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(item)
	defaultName := typ.Name()
	ret = &Resource{
		Name:              func(string) string { return defaultName },
		ID:                columnName(defaultName),
		Typ:               typ,
		Authenticate:      AuthenticateAdmin,
		Pagination:        100000,
		HasModel:          true,
		HasView:           true,
		item:              item,
		StructCache:       cache,
		VisibilityFilter:  defaultVisibilityFilter,
		EditabilityFilter: defaultEditabilityFilter,
		CanCreate:         true,
		CanEdit:           true,
	}

	ret.Actions = map[string]ActionBinder{
		"order":  BindOrder,
		"detail": BindDetail,
		"update": BindUpdate,
		"delete": BindDelete,
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

	ret.admin = a
	a.Resources = append(a.Resources, ret)
	if ret.HasModel {
		_, typFound := a.resourceMap[ret.Typ]
		if typFound {
			return nil, fmt.Errorf("resource with type %s already created", ret.Typ)
		}

		a.resourceMap[ret.Typ] = ret
		a.resourceNameMap[ret.ID] = ret
	}

	return ret, err
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
			render403(request)
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
		return nil, fmt.Errorf("Can't find resource with type %s.", typ)
	}
	return resource, nil
}

func (ar *Resource) db() *sql.DB {
	return ar.admin.getDB()
}

func (ar *Resource) tableName() string {
	return ar.table
}

func (ar *Resource) AddSnippet(template string) {
	ar.Snippets = append(ar.Snippets, Snippet{template})
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

func (ar *Resource) ResourceActionsButtonData(lang string) []ButtonData {
	ret := []ButtonData{}
	if ar.CanCreate {
		ret = append(ret, ButtonData{
			Name: messages.Messages.Get(lang, "admin_new"),
			Url:  fmt.Sprintf("%s/new", ar.ID),
		})
	}

	for _, v := range ar.ResourceActions {
		name := v.Url
		if v.Name != nil {
			name = v.Name(lang)
		}

		ret = append(ret, ButtonData{
			Name: name,
			Url:  fmt.Sprintf("%s/%s", ar.ID, v.Url),
		})
	}

	return ret
}

func (ar *Resource) AddResourceAction(action ResourceAction) {
	ar.ResourceActions = append(ar.ResourceActions, action)
}
