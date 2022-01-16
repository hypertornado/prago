package prago

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type field struct {
	Name        string
	ColumnName  string
	HumanName   func(string) string
	Description func(string) string
	Typ         reflect.Type
	Tags        map[string]string
	fieldOrder  int
	Unique      bool
	CanOrder    bool
	required    bool

	DefaultShow bool

	canView Permission
	canEdit Permission

	resource  *resource
	fieldType *fieldType

	relatedResource *resource
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
	if !field.resource.newResource.getApp().authorize(user, field.resource.newResource.getPermissionView()) {
		return false
	}
	if !field.resource.newResource.getApp().authorize(user, field.canView) {
		return false
	}
	return true
}

func (field *field) authorizeEdit(user *user) bool {
	if !field.authorizeView(user) {
		return false
	}
	if !field.resource.newResource.getApp().authorize(user, field.resource.newResource.getPermissionUpdate()) {
		return false
	}
	if !field.resource.newResource.getApp().authorize(user, field.canEdit) {
		return false
	}
	return true
}

func (resource *resource) newField(f reflect.StructField, order int) *field {
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
		"prago-edit",
		"prago-view",
		"prago-visible",
		"prago-editable",
	} {
		t := f.Tag.Get(v)
		if t != "" {
			panic(fmt.Sprintf("Use of deprecated tag '%s' in field '%s' of resource '%s'", v, ret.Name, ret.resource.newResource.getID()))
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
	if ret.Tags["prago-description"] != "" {
		ret.Description = unlocalized(ret.Tags["prago-description"])
	}

	if ret.Tags["prago-preview"] == "false" {
		ret.DefaultShow = false
	}

	if ret.Tags["prago-preview"] == "false" {
		ret.DefaultShow = false
	}

	if ret.Tags["prago-unique"] == "true" {
		ret.Unique = true
	}

	if ret.Tags["prago-required"] != "" {
		switch ret.Tags["prago-required"] {
		case "true":
			ret.required = true
		case "false":
			break
		default:
			panic(fmt.Sprintf("validating permission 'prago-required' on field '%s' of resource '%s': wrong value '%s'", f.Name, resource.newResource.getName("en"), ret.Tags["prago-required"]))
		}
	}

	if canView := ret.Tags["prago-can-view"]; canView != "" {
		err := resource.newResource.getApp().validatePermission(Permission(canView))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-view' on field '%s' of resource '%s': %s", f.Name, resource.newResource.getName("en"), err))
		}
		ret.canView = Permission(canView)
	}

	if canEdit := ret.Tags["prago-can-edit"]; canEdit != "" {
		err := resource.newResource.getApp().validatePermission(Permission(canEdit))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-edit' on field '%s' of resource '%s': %s", f.Name, resource.newResource.getName("en"), err))
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

	validations := ret.Tags["prago-validations"]
	if validations != "" {
		for _, v := range strings.Split(validations, ",") {
			err := ret.addFieldValidation(v)
			if err != nil {
				panic(fmt.Sprintf("can't add validation on field '%s' of resource '%s': %s", f.Name, resource.newResource.getName("en"), err))
			}
		}
	}

	ret.initFieldType()

	//TODO: better
	if ret.Name != "CreatedAt" && ret.Name != "UpdatedAt" {
		if ret.Typ == reflect.TypeOf(time.Now()) {
			if ret.Tags["prago-type"] == "timestamp" || ret.Name == "CreatedAt" || ret.Name == "UpdatedAt" {
				resource.newResource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.ColumnName)
					if val != "" {
						_, err := time.Parse("2006-01-02 15:04", val)
						if err != nil {
							vc.AddItemError(ret.ColumnName, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			} else {
				resource.newResource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.ColumnName)
					if val != "" {
						_, err := time.Parse("2006-01-02", val)
						if err != nil {
							vc.AddItemError(ret.ColumnName, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			}
		}
	}

	return ret
}

func (field *field) addFieldValidation(nameOfValidation string) error {
	if nameOfValidation == "nonempty" {
		if field.Tags["prago-required"] != "false" {
			field.required = true
		}
		field.resource.newResource.addValidation(func(vc ValidationContext) {
			valid := true
			if field.Typ.Kind() == reflect.Int64 ||
				field.Typ.Kind() == reflect.Int32 ||
				field.Typ.Kind() == reflect.Int ||
				field.Typ.Kind() == reflect.Float64 ||
				field.Typ.Kind() == reflect.Float32 {

				if vc.GetValue(field.ColumnName) == "0" {
					valid = false
				}
			}

			if field.Tags["prago-type"] == "relation" && vc.GetValue(field.ColumnName) == "0" {
				valid = false
			}
			if vc.GetValue(field.ColumnName) == "" {
				valid = false
			}
			if !valid {
				vc.AddItemError(field.ColumnName, messages.Get(vc.Locale(), "admin_validation_not_empty"))
			}
		})
		return nil
	}
	return fmt.Errorf("unknown validation name: %s", nameOfValidation)
}

func (resource *resource) FieldName(nameOfField string, name func(string) string) *resource {
	f := resource.fieldMap[nameOfField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), nameOfField))
	}
	f.HumanName = name
	return resource
}

func (resource *resource) FieldDescription(descriptionOfField string, description func(string) string) *resource {
	f := resource.fieldMap[descriptionOfField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), descriptionOfField))
	}
	f.Description = description
	return resource
}

func (resource *resource) FieldViewTemplate(IDofField string, viewTemplate string) *resource {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), IDofField))
	}
	f.fieldType.viewTemplate = viewTemplate
	return resource
}

func (resource *resource) FieldListCellTemplate(IDofField string, template string) *resource {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), IDofField))
	}
	f.fieldType.listCellTemplate = template
	return resource
}

func (resource *resource) FieldFormTemplate(IDofField string, template string) *resource {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), IDofField))
	}
	f.fieldType.formTemplate = template
	return resource
}

func (resource *resource) FieldDBDescription(IDofField string, description string) *resource {
	f := resource.fieldMap[IDofField]
	if f == nil {
		panic(fmt.Sprintf("can't set field name of resource '%s': field named '%s' not found", resource.newResource.getID(), IDofField))
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

func (field *field) initFieldType() {
	fieldTypes := field.resource.newResource.getApp().fieldTypes
	fieldTypeName := field.Tags["prago-type"]

	ret, found := fieldTypes[fieldTypeName]
	if !found && fieldTypeName != "" {
		panic(fmt.Sprintf("Field type '%s' not found", fieldTypeName))
	}

	if ret == nil {
		ret = &fieldType{}
	}

	if ret.viewTemplate == "" {
		ret.viewTemplate = getDefaultViewTemplate(field.Typ)
	}
	if ret.viewDataSource == nil {
		ret.viewDataSource = getDefaultViewDataSource(field)
	}

	if ret.allowedValues != nil {
		field.resource.newResource.addValidation(func(vc ValidationContext) {
			val := vc.GetValue(field.ColumnName)
			var found bool
			for _, v := range ret.allowedValues {
				if v == val {
					found = true
				}
			}
			if !found {
				vc.AddItemError(field.ColumnName, messages.Get(vc.Locale(), "admin_validation_value"))
			}
		})
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

func (field *field) fieldDescriptionMysql(fieldTypes map[string]*fieldType) string {
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
