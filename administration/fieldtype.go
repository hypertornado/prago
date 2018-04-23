package administration

type FieldType struct {
	ViewTemplate   string
	ViewDataSource func(Resource, User, field, interface{}) interface{}

	DBFieldDescription string

	FormHideLabel  bool
	FormTemplate   string
	FormDataSource func(field, User) interface{}
	FormStringer   func(interface{}) string
}
