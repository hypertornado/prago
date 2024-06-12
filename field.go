package prago

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Field struct {
	fieldClassName  string
	id              string
	name            func(string) string
	nameSetManually bool
	description     func(string) string
	typ             reflect.Type
	tags            map[string]string
	fieldOrder      int
	unique          bool
	canOrder        bool
	required        bool

	defaultHidden bool

	canView Permission
	canEdit Permission

	resource  *Resource
	fieldType *fieldType

	relatedResource *Resource
}

func (resource *Resource) Field(name string) *Field {
	return resource.fieldMap[columnName(name)]
}

func (field *Field) authorizeView(userData UserData) bool {
	if !userData.Authorize(field.resource.canView) {
		return false
	}
	if !userData.Authorize(field.canView) {
		return false
	}
	return true
}

func (field *Field) authorizeEdit(request *Request) bool {
	if !field.authorizeView(request) {
		return false
	}
	if !request.Authorize(field.resource.canUpdate) {
		return false
	}
	if !request.Authorize(field.canEdit) {
		return false
	}
	return true
}

func (resource *Resource) newField(f reflect.StructField, order int) *Field {
	ret := &Field{
		fieldClassName: f.Name,
		id:             columnName(f.Name),
		name:           unlocalized(f.Name),
		typ:            f.Type,
		tags:           make(map[string]string),
		fieldOrder:     order,
		canOrder:       true,

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
			panic(fmt.Sprintf("Use of deprecated tag '%s' in field '%s' of resource '%s'", v, ret.name("en"), ret.resource.getID()))
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
		"prago-icon",
	} {
		ret.tags[v] = f.Tag.Get(v)
	}

	if ret.tags["prago-description"] != "" {
		ret.description = unlocalized(ret.tags["prago-description"])
	}

	if ret.tags["prago-preview"] == "true" {
		ret.defaultHidden = false
	}
	if ret.tags["prago-preview"] == "false" {
		ret.defaultHidden = true
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
			panic(fmt.Sprintf("validating permission 'prago-required' on field '%s' of resource '%s': wrong value '%s'", f.Name, resource.pluralName("en"), ret.tags["prago-required"]))
		}
	}

	if canView := ret.tags["prago-can-view"]; canView != "" {
		err := resource.app.validatePermission(Permission(canView))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-view' on field '%s' of resource '%s': %s", f.Name, resource.pluralName("en"), err))
		}
		ret.canView = Permission(canView)
	}

	if canEdit := ret.tags["prago-can-edit"]; canEdit != "" {
		err := resource.app.validatePermission(Permission(canEdit))
		if err != nil {
			panic(fmt.Sprintf("validating permission 'prago-can-edit' on field '%s' of resource '%s': %s", f.Name, resource.pluralName("en"), err))
		}
		ret.canEdit = Permission(canEdit)
	} else {
		if ret.fieldClassName == "ID" || ret.fieldClassName == "CreatedAt" || ret.fieldClassName == "UpdatedAt" {
			ret.canEdit = nobodyPermission
		}
	}

	name := ret.tags["prago-name"]
	if name != "" {
		ret.nameSetManually = true
		ret.name = unlocalized(name)
	} else {
		nameFunction := getNameFunctionFromStructName(ret.fieldClassName)
		if nameFunction != nil {
			ret.name = nameFunction
		}
	}

	validations := ret.tags["prago-validations"]
	if validations != "" {
		for _, v := range strings.Split(validations, ",") {
			err := ret.addFieldValidation(v)
			if err != nil {
				panic(fmt.Sprintf("can't add validation on field '%s' of resource '%s': %s", f.Name, resource.pluralName("en"), err))
			}
		}
	}

	ret.initFieldType()

	//TODO: better
	if ret.fieldClassName != "CreatedAt" && ret.fieldClassName != "UpdatedAt" {
		if ret.typ == reflect.TypeOf(time.Now()) {
			if ret.tags["prago-type"] == "timestamp" || ret.fieldClassName == "CreatedAt" || ret.fieldClassName == "UpdatedAt" {
				resource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.id)
					if val != "" {
						_, err := time.Parse("2006-01-02 15:04", val)
						if err != nil {
							vc.AddItemError(ret.id, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			} else {
				resource.addValidation(func(vc ValidationContext) {
					val := vc.GetValue(ret.id)
					if val != "" {
						_, err := time.Parse("2006-01-02", val)
						if err != nil {
							vc.AddItemError(ret.id, messages.Get(vc.Locale(), "admin_validation_date_format_error"))
						}
					}
				})
			}
		}
	}

	return ret
}

func getNameFunctionFromStructName(name string) func(string) string {
	id := map[string]string{
		"Name":          "Name",
		"Description":   "Description",
		"Image":         "Image",
		"Hidden":        "Hidden",
		"CreatedAt":     "CreatedAt",
		"UpdatedAt":     "OrderPosition",
		"OrderPosition": "OrderPosition",
		"File":          "File",
		"Place":         "Place",
	}[name]
	if id == "" {
		return nil
	}
	return messages.GetNameFunction(name)
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

				if vc.GetValue(field.id) == "0" {
					valid = false
				}
			}

			if field.tags["prago-type"] == "relation" && vc.GetValue(field.id) == "0" {
				valid = false
			}
			if vc.GetValue(field.id) == "" {
				valid = false
			}
			if !valid {
				vc.AddItemError(field.id, messages.Get(vc.Locale(), "admin_validation_not_empty"))
			}
		})
		return nil
	}
	return fmt.Errorf("unknown validation name: %s", nameOfValidation)
}

func (field *Field) Name(name func(string) string) *Field {
	field.nameSetManually = true
	field.name = name
	return field
}

func (field *Field) getRelatedID() string {
	if field.tags["prago-relation"] != "" {
		return columnName(field.tags["prago-relation"])
	}
	return field.id
}

func (field *Field) GetManuallySetPluralName(locale string) string {
	if !field.nameSetManually {
		return ""
	}
	return field.name(locale)
}

func (field *Field) Description(description func(string) string) *Field {
	field.description = description
	return field
}

func (field *Field) ViewTemplate(template string) *Field {
	field.fieldType.viewTemplate = template
	return field
}

func (field *Field) FormTemplate(template string) *Field {
	field.fieldType.formTemplate = template
	return field
}

func (field *Field) DBDescription(description string) *Field {
	field.fieldType.dbFieldDescription = description
	return field
}

func (field *Field) getIcon() string {
	if field.tags["prago-icon"] != "" {
		return field.tags["prago-icon"]
	}

	if field.fieldType.IsRelation() {
		if field.relatedResource.icon != "" {
			return field.relatedResource.icon
		}
		return iconResource
	}

	if field.fieldType.fieldTypeIcon != "" {
		return field.fieldType.fieldTypeIcon
	}

	if field.id == "id" {
		return "glyphicons-basic-740-hash.svg"
		//return "glyphicons-basic-347-id-badge.svg"
	}

	if field.id == "createdat" || field.id == "updatedat" {
		return "glyphicons-basic-55-clock.svg"
	}

	if field.typ == reflect.TypeOf(true) {
		return "glyphicons-basic-153-square-checkbox.svg"
	}

	if field.typ == reflect.TypeOf("") {
		return "glyphicons-basic-101-text.svg"
	}

	if field.typ.Kind() == reflect.Int || field.typ.Kind() == reflect.Int64 || field.typ.Kind() == reflect.Float64 {
		return "glyphicons-basic-234-calculator.svg"
	}

	if field.typ == reflect.TypeOf(time.Now()) {
		return "glyphicons-basic-46-calendar.svg"
	}
	return ""
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
	fieldTypes := field.resource.app.fieldTypes
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
			val := vc.GetValue(field.id)
			var found bool
			for _, v := range ret.allowedValues {
				if v == val {
					found = true
				}
			}
			if !found {
				vc.AddItemError(field.id, messages.Get(vc.Locale(), "admin_validation_value"))
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
	if field.id == "id" {
		additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
	} else {
		if field.unique {
			additional = "UNIQUE"
		}
	}
	return fmt.Sprintf("`%s` %s %s", field.id, fieldDescription, additional)
}
