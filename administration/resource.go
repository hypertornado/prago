package administration

import (
	"fmt"
	"go/ast"
	"reflect"
	"time"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
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
	actions            []Action
	itemActions        []Action
	relations          []relation
	autoRelations      []relation

	CanView   Permission
	CanEdit   Permission
	CanCreate Permission
	CanDelete Permission
	CanExport Permission

	ActivityLog bool

	PreviewURLFunction func(interface{}) string

	fieldArrays     []*Field
	fieldMap        map[string]*Field
	fieldTypes      map[string]FieldType
	OrderFieldName  string
	OrderColumnName string
}

//CreateResource creates new resource based on item
func (admin *Administration) CreateResource(item interface{}, initFunction func(*Resource)) *Resource {
	if admin.resourcesInited {
		panic("can't create new resource, resources already initiated")
	}

	typ := reflect.TypeOf(item)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("item is not a structure, but " + typ.Kind().String()))
	}

	defaultName := typ.Name()
	ret := &Resource{
		Admin:              admin,
		HumanName:          Unlocalized(defaultName),
		ID:                 columnName(defaultName),
		Typ:                typ,
		ResourceController: admin.AdminController.SubController(),
		ItemsPerPage:       200,
		TableName:          columnName(defaultName),

		CanView:   "",
		CanEdit:   "",
		CanCreate: "",
		CanDelete: "",
		CanExport: "",

		ActivityLog: true,

		fieldMap:   make(map[string]*Field),
		fieldTypes: admin.fieldTypes,
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
		user := GetUser(request)
		if !admin.Authorize(user, resource.CanView) {
			render403(request)
		} else {
			next()
		}
	})

	initResourceActions(admin, resource)
}

//GetURL returns resource url
func (resource Resource) GetURL(suffix string) string {
	ret := resource.Admin.Prefix + "/" + resource.ID
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

func (admin *Administration) getResourceByItem(item interface{}) (*Resource, error) {
	typ := reflect.TypeOf(item).Elem()
	resource, ok := admin.resourceMap[typ]
	if !ok {
		return nil, fmt.Errorf("Can't find resource with type %s.", typ)
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
	return resource.saveItem(db, resource.TableName, item)
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
	return resource.createItem(db, resource.TableName, item)
}

func (resource Resource) newItem(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(resource.Typ))
}

func (resource Resource) newArrayOfItems(item interface{}) {
	reflect.ValueOf(item).Elem().Set(reflect.New(reflect.SliceOf(reflect.PtrTo(resource.Typ))))
}

func (resource Resource) count() int64 {
	var item interface{}
	resource.newItem(&item)
	count, _ := resource.Admin.Query().Count(item)
	return count
}

func (resource Resource) getPaginationData(user User) (ret []ListPaginationData) {
	var ints []int64
	var used bool

	for _, v := range []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000} {
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

		ret = append(ret, ListPaginationData{
			Name:     messages.Messages.ItemsCount(v, user.Locale),
			Value:    v,
			Selected: selected,
		})
	}

	return
}
