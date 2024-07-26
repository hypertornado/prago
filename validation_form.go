package prago

import (
	"errors"
	"html/template"
)

var errValidation = errors.New("validation error")

type FormValidation interface {
	AddError(err string)
	AddItemError(key, err string)
	Valid() bool
	Redirect(string)
	AfterContent(template.HTML)
}

type formValidationData struct {
	RedirectionLocation string
	AfterContent        template.HTML
	Valid               bool
	Errors              []FormValidationError
	ItemErrors          map[string][]FormValidationError
}

type FormValidationError struct {
	Text string
}

type formValidationReport struct {
	Text string
}

func newFormValidationData() *formValidationData {
	ret := &formValidationData{
		Valid: true,
	}
	ret.ItemErrors = map[string][]FormValidationError{}
	return ret
}

type formValidation struct {
	validationData *formValidationData
}

func newFormValidation() *formValidation {
	return &formValidation{
		validationData: newFormValidationData(),
	}
}

func (fv *formValidation) AddError(err string) {

	fv.validationData.Valid = false
	fv.validationData.Errors = append(fv.validationData.Errors, FormValidationError{
		Text: err,
	})
}

func (fv *formValidation) AddItemError(key, err string) {
	fv.validationData.Valid = false
	fv.validationData.ItemErrors[key] = append(fv.validationData.ItemErrors[key], FormValidationError{
		Text: err,
	})
}

func (fv *formValidation) Redirect(url string) {
	fv.validationData.RedirectionLocation = url
}

func (fv *formValidation) AfterContent(content template.HTML) {
	fv.validationData.AfterContent = content
}

func (fv *formValidation) Valid() bool {
	return fv.validationData.Valid
}
