package prago

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type field struct {
	Name       string
	ColumnName string
	HumanName  func(string) string
	Typ        reflect.Type
	Tags       map[string]string
	fieldOrder int
	Unique     bool
	CanOrder   bool

	DefaultShow bool

	canView Permission
	canEdit Permission

	resource  *Resource
	fieldType *FieldType

	relatedResource *Resource
}

//GetRelatedResourceName gets related resource name
func (field field) GetRelatedResourceName() string {
	relatedTag := field.Tags["prago-relation"]
	if relatedTag != "" {
		return strings.ToLower(relatedTag)
	}
	return field.ColumnName
}

func (field *field) authorizeView(user *user) bool {
	if !field.resource.app.authorize(user, field.resource.canView) {
		return false
	}
	if !field.resource.app.authorize(user, field.canView) {
		return false
	}
	return true
}

func (field *field) authorizeEdit(user *user) bool {
	if !field.authorizeView(user) {
		return false
	}
	if !field.resource.app.authorize(user, field.resource.canEdit) {
		return false
	}
	if !field.resource.app.authorize(user, field.canEdit) {
		return false
	}
	return true
}

func (resource *Resource) newField(f reflect.StructField, order int) *field {
	ret := &field{
		Name:        f.Name,
		ColumnName:  columnName(f.Name),
		HumanName:   unlocalized(f.Name),
		Typ:         f.Type,
		Tags:        make(map[string]string),
		fieldOrder:  order,
		CanOrder:    true,
		DefaultShow: false,

		canView: loggedPermission,
		canEdit: loggedPermission,

		resource: resource,
	}

	//remove unused tags
	for _, v := range []string{
		"prago-description",
		"prago-edit",
		"prago-view",
		"prago-visible",
		"prago-editable",
	} {
		t := f.Tag.Get(v)
		if t != "" {
			panic(fmt.Sprintf("Use of deprecated tag '%s' in field '%s' of resource '%s'", v, ret.Name, ret.resource.id))
		}
	}

	for _, v := range []string{
		"prago-can-view",
		"prago-can-edit",

		"prago-name",
		"prago-type",
		"prago-preview",
		"prago-unique",
		"prago-order",
		"prago-order-desc",
		"prago-relation",
	} {
		ret.Tags[v] = f.Tag.Get(v)
	}

	for _, v := range []string{"ID", "Name", "Image", "UpdatedAt"} {
		if ret.Name == v {
			ret.DefaultShow = true
		}
	}
	if ret.Tags["prago-preview"] == "true" {
		ret.DefaultShow = true
	}
	if ret.Tags["prago-preview"] == "false" {
		ret.DefaultShow = false
	}

	if ret.Tags["prago-unique"] == "true" {
		ret.Unique = true
	}

	if canView := ret.Tags["prago-can-view"]; canView != "" {
		err := resource.app.validatePermission(Permission(canView))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-view' on field '%s' of resource '%s': %s", f.Name, resource.name("en"), err))
		}
		ret.canView = Permission(canView)
	}

	if canEdit := ret.Tags["prago-can-edit"]; canEdit != "" {
		err := resource.app.validatePermission(Permission(canEdit))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-edit' on field '%s' of resource '%s': %s", f.Name, resource.name("en"), err))
		}
		ret.canEdit = Permission(canEdit)
	} else {
		if ret.Name == "ID" || ret.Name == "CreatedAt" || ret.Name == "UpdatedAt" {
			ret.canEdit = nobodyPermission
		}
	}

	name := ret.Tags["prago-name"]
	if name != "" {
		ret.HumanName = unlocalized(name)
	} else {
		//TODO: its ugly
		nameFunction := messages.GetNameFunction(ret.Name)
		if nameFunction != nil {
			ret.HumanName = nameFunction
		}
	}

	ret.initFieldType()

	return ret
}

func (resource *Resource) FieldName(nameOfField string, name func(string) string) *Resource {
	f := resource.fieldMap[nameOfField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, nameOfField))
	}
	f.HumanName = name
	return resource
}

func getDefaultStringer(t reflect.Type) func(interface{}) string {
	if reflect.TypeOf(time.Now()) == t {
		return func(i interface{}) string {
			tm := i.(time.Time)
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02")
		}
	}

	switch t.Kind() {
	case reflect.String:
		return func(i interface{}) string {
			return i.(string)
		}
	case reflect.Int64:
		return func(i interface{}) string {
			return fmt.Sprintf("%d", i.(int64))
		}
	case reflect.Float64:
		return func(i interface{}) string {
			return fmt.Sprintf("%f", i.(float64))
		}
	case reflect.Bool:
		return func(i interface{}) string {
			if i.(bool) {
				return "on"
			}
			return ""
		}
	}
	panic("unknown stringer for " + t.String())
}

func getDefaultFormTemplate(t reflect.Type) string {
	if t == reflect.TypeOf(time.Now()) {
		return "admin_item_date"
	}

	switch t.Kind() {
	case reflect.String:
		return "admin_item_input"
	case reflect.Bool:
		return "admin_item_checkbox"
	case reflect.Int64:
		return "admin_item_input_int"
	case reflect.Float64:
		return "admin_item_input_float"
	}
	panic("unknown default form for " + t.String())
}

func (field *field) initFieldType() {
	fieldTypes := field.resource.app.fieldTypes
	fieldTypeName := field.Tags["prago-type"]

	ret, found := fieldTypes[fieldTypeName]
	if !found && fieldTypeName != "" {
		panic(fmt.Sprintf("Field type '%s' not found", fieldTypeName))
	}

	if ret == nil {
		ret = &FieldType{}
	}

	if ret.viewTemplate == "" {
		ret.viewTemplate = getDefaultViewTemplate(field.Typ)
	}
	if ret.viewDataSource == nil {
		ret.viewDataSource = getDefaultViewDataSource(field)
	}

	if ret.formTemplate == "" {
		ret.formTemplate = getDefaultFormTemplate(field.Typ)
	}

	if ret.formStringer == nil {
		ret.formStringer = getDefaultStringer(field.Typ)
	}

	if ret.formTemplate == "admin_item_checkbox" {
		ret.formHideLabel = true
	}

	if ret.listCellDataSource == nil {
		ret.listCellDataSource = ret.viewDataSource
	}
	if ret.listCellTemplate == "" {
		ret.listCellTemplate = ret.viewTemplate
	}
	field.fieldType = ret
}

func (field *field) fieldDescriptionMysql(fieldTypes map[string]*FieldType) string {
	var fieldDescription string

	t, found := fieldTypes[field.Tags["prago-type"]]
	if found && t.dbFieldDescription != "" {
		fieldDescription = t.dbFieldDescription
	} else {
		switch field.Typ.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if field.Typ == dateType {
				if field.Tags["prago-type"] == "date" {
					fieldDescription = "date"
				} else {
					fieldDescription = "datetime"
				}
			}
		case reflect.Bool:
			fieldDescription = "bool NOT NULL"
		case reflect.Float64:
			fieldDescription = "double"
		case reflect.Int64:
			fieldDescription = "bigint(20)"
		case reflect.String:
			if field.Tags["prago-type"] == "text" || field.Tags["prago-type"] == "image" || field.Tags["prago-type"] == "markdown" {
				fieldDescription = "text"
			} else {
				fieldDescription = "varchar(255)"
			}
		default:
			panic("non supported type " + field.Typ.Kind().String())
		}
	}

	additional := ""
	if field.ColumnName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if field.Unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", field.ColumnName, fieldDescription, additional)
}
