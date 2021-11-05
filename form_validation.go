package prago

type FormValidation struct {
	Valid                bool
	RedirectionLocaliton string
	Errors               []FormValidationError
	ItemErrors           map[string][]FormValidationError
}

type FormValidationError struct {
	Text string
}

type Validation func(*Request, *FormValidation)

func (validation *FormValidation) AddError(text string) {
	validation.Valid = false
	validation.Errors = append(validation.Errors, FormValidationError{
		Text: text,
	})
}

func (validation *FormValidation) AddItemError(id, text string) {
	validation.Valid = false
	validation.ItemErrors[id] = append(validation.ItemErrors[id], FormValidationError{
		Text: text,
	})
}

func NewFormValidation() *FormValidation {
	ret := &FormValidation{
		Valid: true,
	}
	ret.ItemErrors = map[string][]FormValidationError{}
	return ret
}
