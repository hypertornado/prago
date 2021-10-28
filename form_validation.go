package prago

type FormValidation struct {
	Valid                bool
	RedirectionLocaliton string
	Errors               []FormValidationError
	ItemErrors           map[string]FormValidationError
}

type FormValidationError struct {
	Text string
}
