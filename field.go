package prago

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
)

type Field struct {
	fieldClassName  string
	id              string
	name            func(string) string
	nameSetManually bool
	description     func(string) string
	useTextOver     bool
	textOver        func(string) string
	typ             reflect.Type
	tags            map[string]string
	fieldOrder      int
	unique          bool
	canOrder        bool
	required        bool

	defaultHidden bool

	isSearchable bool

	canView Permission
	canEdit Permission

	resource  *Resource
	fieldType *fieldType

	relatedResource *Resource

	formContentGenerator func(item *FormItem) template.HTML

	helpURL string

	isWide                 bool
	preventPasswordManager bool

	fixStringValueFN func(string) string

	formFilter    *FormFilter
	suggestionURL string
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

func (field *Field) authorizeEdit(userData UserData) bool {
	if !field.authorizeView(userData) {
		return false
	}
	if !userData.Authorize(field.resource.canUpdate) {
		return false
	}
	if !userData.Authorize(field.canEdit) {
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

		preventPasswordManager: true,
	}

	if ret.id == "name" || ret.id == "description" {
		ret.isSearchable = true
		ret.isWide = true
		ret.preventPasswordManager = true
	}

	//remove unused tags
	for _, v := range []string{
		"prago-edit",
		"prago-view",
		"prago-visible",
		"prago-editable",
		"prago-textover",
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
		"prago-text-over",
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

	if ret.tags["prago-text-over"] == "true" {
		ret.useTextOver = true
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

	ret.initFieldType()

	//TODO: better
	if ret.fieldClassName != "CreatedAt" && ret.fieldClassName != "UpdatedAt" {
		if ret.typ == reflect.TypeOf(time.Now()) {

			resource.addUpdateValidation(func(item any, vc Validation, userData UserData) {
				itemsVal := reflect.ValueOf(item).Elem()
				fieldVal := itemsVal.FieldByName(ret.fieldClassName)
				ivalField := fieldVal.Interface()

				timeVal := ivalField.(time.Time)
				if timeVal.Year() == 0 {
					vc.AddItemError(ret.id, messages.Get(userData.Locale(), "validation_date_format_error"))
				}
			})
		}
	}

	return ret
}

func (resource *Resource) getSearchableFields(request *Request) (ret []*Field) {
	for _, v := range resource.fields {
		if !request.Authorize(v.canView) {
			continue
		}
		if v.isSearchable {
			ret = append(ret, v)
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
		"UpdatedAt":     "UpdatedAt",
		"OrderPosition": "OrderPosition",
		"File":          "File",
		"Place":         "Place",
	}[name]
	if id == "" {
		return nil
	}
	return messages.GetNameFunction(name)
}

func (field *Field) Name(name func(string) string) *Field {
	field.nameSetManually = true
	field.name = name
	return field
}

func (field *Field) HelpURL(url string) *Field {
	field.helpURL = url
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

func (field *Field) TextOver(textOver func(string) string) *Field {
	field.useTextOver = true
	field.textOver = textOver
	return field
}

/*func (field *Field) ViewContentGenerator(fn func(val any) template.HTML) *Field {
	field.viewContentGenerator = fn
	return field
}*/

func (field *Field) FormContentGenerator(fn func(item *FormItem) template.HTML) *Field {
	field.formContentGenerator = fn
	return field
}

func (field *Field) DBDescription(description string) *Field {
	field.fieldType.dbFieldDescription = description
	return field
}

func (field *Field) IsSearchable(isSearchable bool) *Field {
	field.isSearchable = true
	return field
}

func (field *Field) FixStringValue(fn func(string) string) *Field {
	field.fixStringValueFN = fn
	return field
}

func (field *Field) FormFilter(filter *FormFilter) *Field {
	field.formFilter = filter
	return field
}

func (field *Field) SuggestionURL(suggestionURL string) *Field {
	field.suggestionURL = suggestionURL
	return field
}

func (field *Field) Wide(isWide bool) *Field {
	field.isWide = isWide
	return field
}

func (field *Field) PreventPasswordManager(prevent bool) *Field {
	field.preventPasswordManager = prevent
	return field
}

func (field *Field) getIcon() string {
	if field.tags["prago-icon"] != "" {
		return field.tags["prago-icon"]
	}
	if field.fieldType.isRelation() {
		if field.relatedResource.icon != "" {
			return field.relatedResource.icon
		}
	}
	return ""
}

/*func getDefaultStringer(t reflect.Type) func(any) string {
	if reflect.TypeOf(time.Now()) == t {
		return stringerDate
	}

	switch t.Kind() {
	case reflect.String:
		return stringerString
	case reflect.Int64:
		return stringerInt64
	case reflect.Float64:
		return stringerFloat64
	case reflect.Bool:
		return stringerBool
	}
	panic("unknown stringer for " + t.String())
}*/

func stringerString(in any) string {
	return in.(string)
}

func stringerInt64(in any) string {
	return fmt.Sprintf("%d", in.(int64))
}

func stringerFloat64(in any) string {
	return fmt.Sprintf("%f", in.(float64))
}

func stringerBool(in any) string {
	if in.(bool) {
		return "on"
	}
	return ""
}

func stringerDate(in any) string {
	tm := in.(time.Time)
	if tm.IsZero() {
		return ""
	}
	return tm.Format("2006-01-02")
}

func stringerDateTime(in any) string {
	tm := in.(time.Time)
	if tm.IsZero() {
		return ""
	}
	return tm.Format("2006-01-02 15:04")
}

func (field *Field) initFieldType() {
	fieldTypes := field.resource.app.fieldTypes
	fieldTypeName := field.tags["prago-type"]

	if fieldTypeName == "" {
		fieldTypeName = field.typ.Name()
	}

	ret, found := fieldTypes[fieldTypeName]
	if !found {
		panic(fmt.Sprintf("Field type '%s' not found", fieldTypeName))
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
