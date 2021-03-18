package prago

import (
	"fmt"
	"go/ast"
	"reflect"
	"time"
)

//Resource is structure representing one item in admin menu or one table in database
type Resource struct {
	app                 *App
	id                  string
	name                func(locale string) string
	typ                 reflect.Type
	resourceController  *controller
	defaultItemsPerPage int64

	orderByColumn string
	orderDesc     bool

	actions       []*Action
	itemActions   []*Action
	autoRelations []relation

	canView   Permission
	canEdit   Permission
	canCreate Permission
	canDelete Permission
	canExport Permission

	activityLog bool

	previewURL func(interface{}) string

	fieldArrays []*Field
	fieldMap    map[string]*Field
	fieldTypes  map[string]FieldType

	orderFieldName  string
	orderColumnName string
}

//Resource creates new resource based on item
func (app *App) Resource(item interface{} /*, initFunction func(*Resource)*/) *Resource {
	typ := reflect.TypeOf(item)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("item is not a structure, but " + typ.Kind().String()))
	}

	defaultName := typ.Name()
	ret := &Resource{
		app:                 app,
		name:                Unlocalized(defaultName),
		id:                  columnName(defaultName),
		typ:                 typ,
		resourceController:  app.adminController.subController(),
		defaultItemsPerPage: 200,

		canView:   "",
		canEdit:   "",
		canCreate: "",
		canDelete: "",
		canExport: "",

		activityLog: true,

		fieldMap:   make(map[string]*Field),
		fieldTypes: app.fieldTypes,
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := newField(typ.Field(i), i, ret.fieldTypes)
			if field.Tags["prago-type"] == "order" {
				ret.orderFieldName = field.Name
				ret.orderColumnName = field.ColumnName
			}
			ret.fieldArrays = append(ret.fieldArrays, field)
			ret.fieldMap[field.ColumnName] = field
		}
	}

	ret.orderByColumn, ret.orderDesc = ret.getDefaultOrder()

	app.resources = append(app.resources, ret)
	_, typFound := app.resourceMap[ret.typ]
	if typFound {
		panic(fmt.Errorf("resource with type %s already created", ret.typ))
	}

	app.resourceMap[ret.typ] = ret
	app.resourceNameMap[ret.id] = ret

	/*if initFunction != nil {
		initFunction(ret)
	}*/

	return ret
}

func (resource Resource) getItemURL(item interface{}, suffix string) string {
	ret := resource.getURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

//Name sets human name for resource
func (resource *Resource) Name(name func(string) string) *Resource {
	resource.name = name
	return resource
}

//PreviewURLFunction sets function to generate representation of resource item in app
func (resource *Resource) PreviewURLFunction(fn func(interface{}) string) *Resource {
	resource.previewURL = fn
	return resource
}

//ItemsPerPage sets default display value of items per page
func (resource *Resource) ItemsPerPage(itemsPerPage int64) *Resource {
	resource.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *Resource) PermissionView(permission Permission) *Resource {
	resource.canView = permission
	return resource
}

func (resource *Resource) PermissionEdit(permission Permission) *Resource {
	resource.canEdit = permission
	return resource
}

func (resource *Resource) PermissionCreate(permission Permission) *Resource {
	resource.canCreate = permission
	return resource
}

func (resource *Resource) PermissionDelete(permission Permission) *Resource {
	resource.canDelete = permission
	return resource
}

func (resource *Resource) PermissionExport(permission Permission) *Resource {
	resource.canExport = permission
	return resource
}

func (app *App) getResourceByName(name string) *Resource {
	return app.resourceNameMap[columnName(name)]
}

func (app *App) initResources() {
	for _, v := range app.resources {
		app.initResource(v)
	}

}

func (app *App) initResource(resource *Resource) {
	resource.resourceController.addAroundAction(func(request Request, next func()) {
		user := request.getUser()
		if !app.authorize(user, resource.canView) {
			render403(request)
		} else {
			next()
		}
	})

	initResourceActions(resource)
	initResourceAPIs(resource)
}

func (resource Resource) getURL(suffix string) string {
	url := resource.id
	if len(suffix) > 0 {
		url += "/" + suffix
	}
	return resource.app.getAdminURL(url)
}

func (app *App) getResourceByItem(item interface{}) (*Resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := app.resourceMap[typ]
	if !ok {
		return nil, fmt.Errorf("can't find resource with type %s", typ)
	}
	return resource, nil
}

func (resource Resource) saveWithDBIface(item interface{}, db dbIface) error {
	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() &&
		val.FieldByName(fn).CanSet() &&
		val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}
	return resource.saveItem(db, resource.id, item)
}

func (resource Resource) createWithDBIface(item interface{}, db dbIface) error {
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
	return resource.createItem(db, resource.id, item)
}

func (resource Resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(resource.typ))
}

func (resource Resource) newArrayOfItems(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(reflect.SliceOf(reflect.PtrTo(resource.typ))))
}

func (resource Resource) count() int64 {
	var item interface{}
	resource.newItem(&item)
	count, _ := resource.app.Query().Count(item)
	return count
}

func (resource Resource) cachedCountName() string {
	return fmt.Sprintf("resource_count-%s", resource.id)
}

func (resource Resource) getCachedCount() int64 {
	return resource.app.cache.Load(resource.cachedCountName(), func() interface{} {
		return resource.count()
	}).(int64)
}

func (resource Resource) updateCachedCount() error {
	return resource.app.cache.Set(resource.cachedCountName(), resource.count())
}

func (resource Resource) getPaginationData(user User) (ret []listPaginationData) {
	var ints []int64
	var used bool

	for _, v := range []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000} {
		if !used {
			if v == resource.defaultItemsPerPage {
				used = true
			}
			if resource.defaultItemsPerPage < v {
				used = true
				ints = append(ints, resource.defaultItemsPerPage)
			}
		}
		ints = append(ints, v)
	}

	if resource.defaultItemsPerPage > ints[len(ints)-1] {
		ints = append(ints, resource.defaultItemsPerPage)
	}

	for _, v := range ints {
		var selected bool
		if v == resource.defaultItemsPerPage {
			selected = true
		}

		ret = append(ret, listPaginationData{
			Name:     messages.ItemsCount(v, user.Locale),
			Value:    v,
			Selected: selected,
		})
	}

	return
}
