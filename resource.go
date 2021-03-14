package prago

import (
	"fmt"
	"go/ast"
	"reflect"
	"time"
)

//Resource is structure representing one item in admin menu or one table in database
type Resource struct {
	app                *App
	id                 string
	HumanName          func(locale string) string
	typ                reflect.Type
	resourceController *Controller
	ItemsPerPage       int64
	OrderByColumn      string
	OrderDesc          bool
	//TableName          string
	actions       []*Action
	itemActions   []*Action
	relations     []relation
	autoRelations []relation

	CanView   Permission
	CanEdit   Permission
	CanCreate Permission
	CanDelete Permission
	CanExport Permission

	ActivityLog bool

	PreviewURLFunction func(interface{}) string

	fieldArrays []*Field
	fieldMap    map[string]*Field
	fieldTypes  map[string]FieldType

	OrderFieldName  string
	OrderColumnName string
}

//CreateResource creates new resource based on item
func (app *App) CreateResource(item interface{}, initFunction func(*Resource)) *Resource {
	typ := reflect.TypeOf(item)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("item is not a structure, but " + typ.Kind().String()))
	}

	defaultName := typ.Name()
	ret := &Resource{
		app:                app,
		HumanName:          Unlocalized(defaultName),
		id:                 columnName(defaultName),
		typ:                typ,
		resourceController: app.adminController.subController(),
		ItemsPerPage:       200,
		//TableName:          columnName(defaultName),

		CanView:   "",
		CanEdit:   "",
		CanCreate: "",
		CanDelete: "",
		CanExport: "",

		ActivityLog: true,

		fieldMap:   make(map[string]*Field),
		fieldTypes: app.fieldTypes,
	}

	for i := 0; i < typ.NumField(); i++ {
		if ast.IsExported(typ.Field(i).Name) {
			field := newField(typ.Field(i), i, ret.fieldTypes)
			if field.Tags["prago-type"] == "order" {
				ret.OrderFieldName = field.Name
				ret.OrderColumnName = field.ColumnName
			}
			ret.fieldArrays = append(ret.fieldArrays, field)
			ret.fieldMap[field.ColumnName] = field
		}
	}

	ret.OrderByColumn, ret.OrderDesc = ret.getDefaultOrder()

	app.resources = append(app.resources, ret)
	_, typFound := app.resourceMap[ret.typ]
	if typFound {
		panic(fmt.Errorf("resource with type %s already created", ret.typ))
	}

	app.resourceMap[ret.typ] = ret
	app.resourceNameMap[ret.id] = ret

	if initFunction != nil {
		initFunction(ret)
	}

	app.initResource(ret)

	return ret
}

//GetItemURL gets item url
func (resource Resource) GetItemURL(item interface{}, suffix string) string {
	ret := resource.getURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func (app *App) getResourceByName(name string) *Resource {
	return app.resourceNameMap[columnName(name)]
}

func (app *App) initResource(resource *Resource) {

	resource.resourceController.AddAroundAction(func(request Request, next func()) {
		user := request.GetUser()
		if !app.Authorize(user, resource.CanView) {
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
	return resource.app.GetAdminURL(url)
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
	return resource.app.Cache.Load(resource.cachedCountName(), func() interface{} {
		return resource.count()
	}).(int64)
}

func (resource Resource) updateCachedCount() error {
	return resource.app.Cache.Set(resource.cachedCountName(), resource.count())
}

func (resource Resource) getPaginationData(user User) (ret []listPaginationData) {
	var ints []int64
	var used bool

	for _, v := range []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000} {
		if !used {
			if v == resource.ItemsPerPage {
				used = true
			}
			if resource.ItemsPerPage < v {
				used = true
				ints = append(ints, resource.ItemsPerPage)
			}
		}
		ints = append(ints, v)
	}

	if resource.ItemsPerPage > ints[len(ints)-1] {
		ints = append(ints, resource.ItemsPerPage)
	}

	for _, v := range ints {
		var selected bool
		if v == resource.ItemsPerPage {
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
