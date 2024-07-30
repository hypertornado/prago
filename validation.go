package prago

type Validation interface {
	AddError(err string)
	AddItemError(key, err string)
	Valid() bool
}

type ValidationError struct {
	Field string
	Text  string
}

func (resource *Resource) validateUpdate(item any, user UserData) *itemValidation {
	itemValidation := newItemValidation()
	for _, validation := range resource.updateValidations {
		validation(item, itemValidation, user)
	}
	return itemValidation
}

func (resource *Resource) validateDelete(item any, user UserData) *itemValidation {
	itemValidation := newItemValidation()
	for _, validation := range resource.deleteValidations {
		validation(item, itemValidation, user)
	}
	return itemValidation
}

func TestValidationUpdate[T any](app *App, item *T, user UserData) ([]ValidationError, bool) {
	resource := getResource[T](app)
	validation := resource.validateUpdate(item, user)
	return validation.errors, validation.Valid()
}

func TestValidationDelete[T any](app *App, item *T, user UserData) ([]ValidationError, bool) {
	resource := getResource[T](app)
	validation := resource.validateDelete(item, user)
	return validation.errors, validation.Valid()
}
