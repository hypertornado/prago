package prago

import (
	"fmt"
	"strings"
)

type itemValidation struct {
	valid  bool
	errors []ValidationError
}

func newItemValidation() *itemValidation {
	ret := &itemValidation{
		valid: true,
	}
	return ret
}

func (iv *itemValidation) AddError(err string) {
	iv.valid = false
	iv.errors = append(iv.errors, ValidationError{
		Text: err,
	})
}

func (iv *itemValidation) AddItemError(key, err string) {
	iv.valid = false
	iv.errors = append(iv.errors, ValidationError{
		Field: key,
		Text:  err,
	})
}

func (iv *itemValidation) Valid() bool {
	return iv.valid
}

func (iv *itemValidation) TextErrorReport(id int64, locale string) formValidationReport {
	var errors []string
	for _, v := range iv.errors {
		if v.Field != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", v.Field, v.Text))
		} else {
			errors = append(errors, v.Text)
		}
	}

	return formValidationReport{
		Text: fmt.Sprintf("%s (id %d): %s",
			messages.Get(locale, "admin_validation_error"),
			id,
			strings.Join(errors, "; "),
		),
	}

}
