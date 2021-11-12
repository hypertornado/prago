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

type ValidationContext interface {
	Locale() string
	GetValue(key string) string
	GetValues(key string) []string
	AddError(err string)
	AddItemError(key, err string)
	Validation() *FormValidation
	Valid() bool
}

type Validation func(ValidationContext)

type requestValidation struct {
	request    *Request
	validation *FormValidation
}

func newRequestValidation(request *Request) *requestValidation {
	return &requestValidation{
		request:    request,
		validation: NewFormValidation(),
	}
}

func (rv *requestValidation) Locale() string {
	return rv.request.user.Locale
}

func (rv *requestValidation) GetValue(key string) string {
	return rv.request.Params().Get(key)
}

func (rv *requestValidation) GetValues(key string) []string {
	return rv.request.Params()[key]
}

func (rv *requestValidation) AddError(err string) {
	rv.validation.AddError(err)
}

func (rv *requestValidation) AddItemError(key, err string) {
	rv.validation.AddItemError(key, err)
}

func (rv *requestValidation) Validation() *FormValidation {
	return rv.validation
}

func (rv *requestValidation) Valid() bool {
	return rv.validation.Valid

}
