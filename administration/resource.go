package administration

import (
	"fmt"
	"github.com/hypertornado/prago"
	"reflect"
	"time"
)

//Resource is structure representing one item in admin menu or one table in database
type Resource struct {
	Admin              *Administration
	ID                 string
	HumanName          func(locale string) string
	Typ                reflect.Type
	ResourceController *prago.Controller
	ItemsPerPage       int64
	OrderByColumn      string
	OrderDesc          bool
	TableName          string
	StructCache        *structCache
	AfterFormCreated   func(f *Form, request prago.Request, newItem bool) *Form //TODO: remove this
	visibilityFilter   structFieldFilter
	editabilityFilter  structFieldFilter
	actions            []Action
	itemActions        []Action
	relations          []relation

	CanView   Permission
	CanEdit   Permission
	CanCreate Permission
	CanDelete Permission
	CanExport Permission

	ActivityLog bool

	PreviewURLFunction func(interface{}) string
}

func (resource Resource) Field(name string) *field {
	if resource.StructCache == nil {
		return nil
	}
	return resource.StructCache.fieldMap[name]
}

func (resource Resource) GetURL(suffix string) string {
	ret := resource.Admin.Prefix + "/" + resource.ID
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//CreateResource creates new resource based on item
func (admin *Administration) CreateResource(item interface{}, initFunction func(*Resource)) *Resource {
	cache, err := newStructCache(item, admin.fieldTypes)
	must(err)

	typ := reflect.TypeOf(item)
	defaultName := typ.Name()
	ret := &Resource{
		Admin:              admin,
		HumanName:          Unlocalized(defaultName),
		ID:                 columnName(defaultName),
		Typ:                typ,
		ResourceController: admin.AdminController.SubController(),
		ItemsPerPage:       200,
		TableName:          columnName(defaultName),
		StructCache:        cache,
		visibilityFilter:   defaultVisibilityFilter,
		editabilityFilter:  defaultEditabilityFilter,

		CanView:   "",
		CanEdit:   "",
		CanCreate: "",
		CanDelete: "",
		CanExport: "",

		ActivityLog: true,
	}

	ret.OrderByColumn, ret.OrderDesc = cache.GetDefaultOrder()

	admin.Resources = append(admin.Resources, ret)
	_, typFound := admin.resourceMap[ret.Typ]
	if typFound {
		panic(fmt.Errorf("resource with type %s already created", ret.Typ))
	}

	admin.resourceMap[ret.Typ] = ret
	admin.resourceNameMap[ret.ID] = ret

	if initFunction != nil {
		initFunction(ret)
	}

	return ret
}

func (admin *Administration) initResource(resource *Resource) {
	resource.ResourceController.AddAroundAction(func(request prago.Request, next func()) {
		request.SetData("admin_resource", resource)
		user := GetUser(request)
		if !admin.Authorize(user, resource.CanView) {
			render403(request)
		} else {
			next()
		}
	})

	initResourceActions(admin, resource)
}

func (a *Administration) getResourceByItem(item interface{}) (*Resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := a.resourceMap[typ]
	if !ok {
		return nil, fmt.Errorf("Can't find resource with type %s.", typ)
	}
	return resource, nil
}

func (resource *Resource) unsafeDropTable() error {
	return dropTable(resource.Admin.db, resource.TableName)
}

func (resource *Resource) migrate(verbose bool) error {
	_, err := getTableDescription(resource.Admin.db, resource.TableName)
	if err == nil {
		return migrateTable(resource.Admin.db, resource.TableName, resource.StructCache, verbose)
	}
	return createTable(resource.Admin.db, resource.TableName, resource.StructCache, verbose)
}

func (resource *Resource) saveWithDBIface(item interface{}, db dbIface) error {
	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() && val.FieldByName(fn).CanSet() && val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}

	return resource.StructCache.saveItem(db, resource.TableName, item)
}

func (resource *Resource) createWithDBIface(item interface{}, db dbIface) error {
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
	return resource.StructCache.createItem(db, resource.TableName, item)
}

func (resource *Resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(resource.Typ))
}

func (resource *Resource) newArrayOfItems(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(reflect.SliceOf(reflect.PtrTo(resource.Typ))))
}
