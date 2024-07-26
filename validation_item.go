package prago

import (
	"fmt"
	"strings"
)

type ItemValidation interface {
	AddError(err string)
	AddItemError(key, err string)
	Valid() bool
}

type itemValidation struct {
	valid      bool
	errors     []FormValidationError
	itemErrors map[string][]FormValidationError
}

func newItemValidation() *itemValidation {
	ret := &itemValidation{
		valid: true,
	}
	ret.itemErrors = map[string][]FormValidationError{}
	return ret
}

func (iv *itemValidation) AddError(err string) {
	iv.valid = false
	iv.errors = append(iv.errors, FormValidationError{
		Text: err,
	})
}

func (iv *itemValidation) AddItemError(key, err string) {
	iv.valid = false
	iv.itemErrors[key] = append(iv.itemErrors[key], FormValidationError{
		Text: err,
	})
}

func (iv *itemValidation) Valid() bool {
	return iv.valid
}

func (iv *itemValidation) TextErrorReport(id int64, locale string) formValidationReport {
	var errors []string
	for _, v := range iv.errors {
		errors = append(errors, v.Text)
	}

	for k, v := range iv.itemErrors {
		for _, v2 := range v {
			errors = append(errors, fmt.Sprintf("%s: %s", k, v2.Text))
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
