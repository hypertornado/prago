package prago

type FormValidation struct {
	//Valid                bool
	RedirectionLocaliton string
	Errors               []FormValidationError
	ItemErrors           map[string][]FormValidationError
}

type FormValidationError struct {
	Text string
}

func (validation *FormValidation) AddError(text string) {
	validation.Errors = append(validation.Errors, FormValidationError{
		Text: text,
	})
}

func (validation *FormValidation) AddItemError(id, text string) {
	validation.ItemErrors[id] = append(validation.ItemErrors[id], FormValidationError{
		Text: text,
	})
}

func NewFormValidation() *FormValidation {
	ret := &FormValidation{}
	ret.ItemErrors = map[string][]FormValidationError{}
	return ret
}
