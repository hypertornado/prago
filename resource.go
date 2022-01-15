package prago

import (
	"fmt"
	"go/ast"
	"reflect"
	"time"
)

//Resource is structure representing one item in admin menu or one table in database
type resource struct {
	app                 *App
	id                  string
	name                func(locale string) string
	typ                 reflect.Type
	resourceController  *controller
	defaultItemsPerPage int64

	orderByColumn string
	orderDesc     bool

	actions     []*Action
	itemActions []*Action

	relations []relation

	canView   Permission
	canCreate Permission
	canUpdate Permission
	canDelete Permission
	canExport Permission

	activityLog bool

	previewURL func(interface{}) string

	fieldArrays []*field
	fieldMap    map[string]*field

	validations       []Validation
	deleteValidations []Validation

	orderField *field
}

//Resource creates new resource based on item
func (app *App) oldNewResource(item interface{}) *resource {
	typ := reflect.TypeOf(item)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("item is not a structure, but " + typ.Kind().String()))
	}

	defaultName := typ.Name()
	ret := &resource{
		app:                 app,
		name:                unlocalized(defaultName),
		id:                  columnName(defaultName),
		typ:                 typ,
		resourceController:  app.adminController.subController(),
		defaultItemsPerPage: 200,

		canView:   sysadminPermission,
		canCreate: loggedPermission,
		canUpdate: loggedPermission,
		canDelete: loggedPermission,
		canExport: loggedPermission,

		activityLog: true,

		fieldMap: make(map[string]*field),
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := ret.newField(typ.Field(i), i)
			if field.Tags["prago-type"] == "order" {
				ret.orderField = field
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

	app.initResource(ret)

	return ret
}

func (resource resource) allowsMultipleActions(user *user) (ret bool) {
	if resource.app.authorize(user, resource.canDelete) {
		ret = true
	}
	if resource.app.authorize(user, resource.canUpdate) {
		ret = true
	}
	return ret
}

func (resource resource) getMultipleActions(user *user) (ret []listMultipleAction) {
	if !resource.allowsMultipleActions(user) {
		return nil
	}

	if resource.app.authorize(user, resource.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if resource.app.authorize(user, resource.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	if resource.app.authorize(user, resource.canDelete) {
		ret = append(ret, listMultipleAction{
			ID:       "delete",
			Name:     "Smazat",
			IsDelete: true,
		})
	}
	ret = append(ret, listMultipleAction{
		ID:   "cancel",
		Name: "Storno",
	})
	return
}

func (resource resource) getItemURL(item interface{}, suffix string) string {
	ret := resource.getURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

//PreviewURLFunction sets function to generate representation of resource item in app
func (resource *resource) PreviewURLFunction(fn func(interface{}) string) *resource {
	resource.previewURL = fn
	return resource
}

//ItemsPerPage sets default display value of items per page
func (resource *resource) ItemsPerPage(itemsPerPage int64) *resource {
	resource.defaultItemsPerPage = itemsPerPage
	return resource
}

func (resource *resource) PermissionView(permission Permission) *resource {
	must(resource.app.validatePermission(permission))
	resource.canView = permission
	return resource
}

//PermissionUpdate sets permission to edit functions, if there is no create and delete permissions set, it set them too
func (resource *resource) PermissionUpdate(permission Permission) *resource {
	must(resource.app.validatePermission(permission))
	if resource.canCreate == loggedPermission {
		resource.canCreate = permission
	}
	if resource.canDelete == loggedPermission {
		resource.canDelete = permission
	}
	resource.canUpdate = permission
	return resource
}

func (resource *resource) PermissionCreate(permission Permission) *resource {
	must(resource.app.validatePermission(permission))
	resource.canCreate = permission
	return resource
}

func (resource *resource) PermissionDelete(permission Permission) *resource {
	must(resource.app.validatePermission(permission))
	resource.canDelete = permission
	return resource
}

func (resource *resource) PermissionExport(permission Permission) *resource {
	must(resource.app.validatePermission(permission))
	resource.canExport = permission
	return resource
}

func (resource *resource) Validation(validation Validation) *resource {
	resource.validations = append(resource.validations, validation)
	return resource
}

func (resource *resource) DeleteValidation(validation Validation) *resource {
	resource.deleteValidations = append(resource.deleteValidations, validation)
	return resource
}

func (app *App) getResourceByName(name string) *resource {
	return app.resourceNameMap[columnName(name)]
}

func (app *App) initResource(resource *resource) {
	resource.resourceController.addAroundAction(func(request *Request, next func()) {
		if !app.authorize(request.user, resource.canView) {
			render403(request)
		} else {
			next()
		}
	})
}

func (app *App) initDefaultResourceActions() {
	for _, v := range app.resources {
		initDefaultResourceActions(v)
		initDefaultResourceAPIs(v)
	}
}

func (resource resource) getURL(suffix string) string {
	url := resource.id
	if len(suffix) > 0 {
		url += "/" + suffix
	}
	return resource.app.getAdminURL(url)
}

func (app *App) getResourceByItem(item interface{}) (*resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := app.resourceMap[typ]
	if !ok {
		return nil, fmt.Errorf("can't find resource with type %s", typ)
	}
	return resource, nil
}

func (resource resource) saveWithDBIface(item interface{}, db dbIface, debugSQL bool) error {
	val := reflect.ValueOf(item).Elem()
	timeVal := reflect.ValueOf(time.Now())
	fn := "UpdatedAt"
	if val.FieldByName(fn).IsValid() &&
		val.FieldByName(fn).CanSet() &&
		val.FieldByName(fn).Type() == timeVal.Type() {
		val.FieldByName(fn).Set(timeVal)
	}
	return resource.saveItem(db, resource.id, item, debugSQL)
}

func (resource resource) createWithDBIface(item interface{}, db dbIface, debugSQL bool) error {
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
	return resource.createItem(db, resource.id, item, debugSQL)
}

func (resource resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(resource.typ))
}

func (resource resource) count() int64 {
	count, _ := resource.query().count()
	return count
}

func (resource resource) cachedCountName() string {
	return fmt.Sprintf("resource_count-%s", resource.id)
}

func (resource resource) getCachedCount() int64 {
	return resource.app.cache.Load(resource.cachedCountName(), func() interface{} {
		return resource.count()
	}).(int64)
}

func (resource resource) updateCachedCount() error {
	return resource.app.cache.set(resource.cachedCountName(), resource.count())
}

func (resource resource) getPaginationData(user *user) (ret []listPaginationData) {
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
