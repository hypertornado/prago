package prago

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Field struct {
	name        string
	columnName  string
	humanName   func(string) string
	description func(string) string
	typ         reflect.Type
	tags        map[string]string
	fieldOrder  int
	unique      bool
	canOrder    bool
	required    bool

	defaultShow bool

	canView Permission
	canEdit Permission

	resource  resourceIface
	fieldType *fieldType

	relatedResource resourceIface
}

func (resource *Resource[T]) Field(name string) *Field {
	return resource.fieldMap[columnName(name)]
}

//GetRelatedResourceName gets related resource name
func (field *Field) GetRelatedResourceName() string {
	relatedTag := field.tags["prago-relation"]
	if relatedTag != "" {
		return strings.ToLower(relatedTag)
	}
	return field.columnName
}

func (field *Field) authorizeView(user *user) bool {
	if !field.resource.getApp().authorize(user, field.resource.getPermissionView()) {
		return false
	}
	if !field.resource.getApp().authorize(user, field.canView) {
		return false
	}
	return true
}

func (field *Field) authorizeEdit(user *user) bool {
	if !field.authorizeView(user) {
		return false
	}
	if !field.resource.getApp().authorize(user, field.resource.getPermissionUpdate()) {
		return false
	}
	if !field.resource.getApp().authorize(user, field.canEdit) {
		return false
	}
	return true
}

func (resource *Resource[T]) newField(f reflect.StructField, order int) *Field {
	ret := &Field{
		name:        f.Name,
		columnName:  columnName(f.Name),
		humanName:   unlocalized(f.Name),
		typ:         f.Type,
		tags:        make(map[string]string),
		fieldOrder:  order,
		canOrder:    true,
		defaultShow: false,

		canView: loggedPermission,
		canEdit: loggedPermission,

		resource: resource,
	}

	//remove unused tags
	for _, v := range []string{
		"prago-edit",
		"prago-view",
		"prago-visible",
		"prago-editable",
	} {
		t := f.Tag.Get(v)
		if t != "" {
			panic(fmt.Sprintf("Use of deprecated tag '%s' in field '%s' of resource '%s'", v, ret.name, ret.resource.getID()))
		}
	}

	for _, v := range []string{
		"prago-can-view",
		"prago-can-edit",
		"prago-name",
		"prago-description",
		"prago-type",
		"prago-preview",
		"prago-unique",
		"prago-order",
		"prago-order-desc",
		"prago-relation",
		"prago-validations",
		"prago-required",
	} {
		ret.tags[v] = f.Tag.Get(v)
	}

	for _, v := range []string{"ID", "Name", "Image", "UpdatedAt"} {
		if ret.name == v {
			ret.defaultShow = true
		}
	}

	if ret.tags["prago-description"] != "" {
		ret.description = unlocalized(ret.tags["prago-description"])
	}

	if ret.tags["prago-preview"] == "true" {
		ret.defaultShow = true
	}
	if ret.tags["prago-preview"] == "false" {
		ret.defaultShow = false
	}

	if ret.tags["prago-unique"] == "true" {
		ret.unique = true
	}

	if ret.tags["prago-required"] != "" {
		switch ret.tags["prago-required"] {
		case "true":
			ret.required = true
		case "false":
			break
		default:
			panic(fmt.Sprintf("validating permission 'prago-required' on field '%s' of resource '%s': wrong value '%s'", f.Name, resource.getName("en"), ret.tags["prago-required"]))
		}
	}

	if canView := ret.tags["prago-can-view"]; canView != "" {
		err := resource.getApp().validatePermission(Permission(canView))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-view' on field '%s' of resource '%s': %s", f.Name, resource.getName("en"), err))
		}
		ret.canView = Permission(canView)
	}

	if canEdit := ret.tags["prago-can-edit"]; canEdit != "" {
		err := resource.getApp().validatePermission(Permission(canEdit))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-edit' on field '%s' of resource '%s': %s", f.Name, resource.getName("en"), err))
		}
		ret.canEdit = Permission(canEdit)
	} else {
		if ret.name == "ID" || ret.name == "CreatedAt" || ret.name == "UpdatedAt" {
			ret.canEdit = nobodyPermission
		}
	}

	name := ret.tags["prago-name"]
	if name != "" {
		ret.humanName = unlocalized(name)
	} else {
		//TODO: its ugly
		nameFunction := messages.GetNameFunction(ret.name)
		if nameFunction != nil {
			ret.humanName = nameFunction
		}
	}

	validations := ret.tags["prago-validations"]
	if validations != "" {
		for _, v := range strings.Split(validations, ",") {
			err := ret.addFieldValidation(v)
			if err != nil {
				panic(fmt.Sprintf("can't add validation on field '%s' of resource '%s': %s", f.Name, resource.getName("en"), err))
			}
		}
	}

	ret.initFieldType()

	//TODO: better
	if ret.name != "CreatedAt" && ret.name != "UpdatedAt" {
		if ret.typ == reflect.TypeOf(time.Now()) {
			if ret.tags["prago-type"] == "timestamp" || ret.name == "CreatedAt" || ret.name == "UpdatedAt" {
				resource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.columnName)
					if val != "" {
						_, err := time.Parse("2006-01-02 15:04", val)
						if err != nil {
							vc.AddItemError(ret.columnName, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			} else {
				resource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.columnName)
					if val != "" {
						_, err := time.Parse("2006-01-02", val)
						if err != nil {
							vc.AddItemError(ret.columnName, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			}
		}
	}

	return ret
}

func (field *Field) addFieldValidation(nameOfValidation string) error {
	if nameOfValidation == "nonempty" {
		if field.tags["prago-required"] != "false" {
			field.required = true
		}
		field.resource.addValidation(func(vc ValidationContext) {
			valid := true
			if field.typ.Kind() == reflect.Int64 ||
				field.typ.Kind() == reflect.Int32 ||
				field.typ.Kind() == reflect.Int ||
				field.typ.Kind() == reflect.Float64 ||
				field.typ.Kind() == reflect.Float32 {

				if vc.GetValue(field.columnName) == "0" {
					valid = false
				}
			}

			if field.tags["prago-type"] == "relation" && vc.GetValue(field.columnName) == "0" {
				valid = false
			}
			if vc.GetValue(field.columnName) == "" {
				valid = false
			}
			if !valid {
				vc.AddItemError(field.columnName, messages.Get(vc.Locale(), "admin_validation_not_empty"))
			}
		})
		return nil
	}
	return fmt.Errorf("unknown validation name: %s", nameOfValidation)
}

func (field *Field) Name(name func(string) string) *Field {
	/*f := resource.fieldMap[nameOfField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, nameOfField))
	}*/
	field.humanName = name
	return field
}

func (resource *Resource[T]) FieldDescription(descriptionOfField string, description func(string) string) *Resource[T] {
	f := resource.fieldMap[descriptionOfField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, descriptionOfField))
	}
	f.description = description
	return resource
}

func (resource *Resource[T]) FieldViewTemplate(IDofField string, viewTemplate string) *Resource[T] {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, IDofField))
	}
	f.fieldType.viewTemplate = viewTemplate
	return resource
}

func (resource *Resource[T]) FieldListCellTemplate(IDofField string, template string) *Resource[T] {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, IDofField))
	}
	f.fieldType.listCellTemplate = template
	return resource
}

func (resource *Resource[T]) FieldFormTemplate(IDofField string, template string) *Resource[T] {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, IDofField))
	}
	f.fieldType.formTemplate = template
	return resource
}

func (resource *Resource[T]) FieldDBDescription(IDofField string, description string) *Resource[T] {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.id, IDofField))
	}
	f.fieldType.dbFieldDescription = description
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

func (field *Field) initFieldType() {
	fieldTypes := field.resource.getApp().fieldTypes
	fieldTypeName := field.tags["prago-type"]

	ret, found := fieldTypes[fieldTypeName]
	if !found && fieldTypeName != "" {
		panic(fmt.Sprintf("Field type '%s' not found", fieldTypeName))
	}

	if ret == nil {
		ret = &fieldType{}
	}

	if ret.viewTemplate == "" {
		ret.viewTemplate = getDefaultViewTemplate(field.typ)
	}
	if ret.viewDataSource == nil {
		ret.viewDataSource = getDefaultViewDataSource(field)
	}

	if ret.allowedValues != nil {
		field.resource.addValidation(func(vc ValidationContext) {
			val := vc.GetValue(field.columnName)
			var found bool
			for _, v := range ret.allowedValues {
				if v == val {
					found = true
				}
			}
			if !found {
				vc.AddItemError(field.columnName, messages.Get(vc.Locale(), "admin_validation_value"))
			}
		})
	}

	if ret.formTemplate == "" {
		ret.formTemplate = getDefaultFormTemplate(field.typ)
	}

	if ret.formStringer == nil {
		ret.formStringer = getDefaultStringer(field.typ)
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

func (field *Field) fieldDescriptionMysql(fieldTypes map[string]*fieldType) string {
	var fieldDescription string

	t, found := fieldTypes[field.tags["prago-type"]]
	if found && t.dbFieldDescription != "" {
		fieldDescription = t.dbFieldDescription
	} else {
		switch field.typ.Kind() {
		case reflect.Struct:
			dateType := reflect.TypeOf(time.Now())
			if field.typ == dateType {
				if field.tags["prago-type"] == "date" {
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
			if field.tags["prago-type"] == "text" || field.tags["prago-type"] == "image" || field.tags["prago-type"] == "markdown" {
				fieldDescription = "text"
			} else {
				fieldDescription = "varchar(255)"
			}
		default:
			panic("non supported type " + field.typ.Kind().String())
		}
	}

	additional := ""
	if field.columnName == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if field.unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("%s %s %s", field.columnName, fieldDescription, additional)
}
