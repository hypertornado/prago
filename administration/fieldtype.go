package administration

type FieldType struct {
	ViewTemplate   string
	ViewDataSource *func(User) interface{}

	DBFieldDescription string

	FormHideLabel  bool
	FormTemplate   string
	FormDataSource func(field, User) interface{}
	FormStringer   func(interface{}) string
}
