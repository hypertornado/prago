package prago

import (
	"html/template"
)

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
	//Valid               bool
	Errors []ValidationError
}

type formValidationReport struct {
	Text string
}

func newFormValidationData() *formValidationData {
	ret := &formValidationData{
		//Valid: true,
	}
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

	//fv.validationData.Valid = false
	fv.validationData.Errors = append(fv.validationData.Errors, ValidationError{
		Text: err,
	})
}

func (fv *formValidation) AddItemError(key, err string) {
	//fv.validationData.Valid = false
	fv.validationData.Errors = append(fv.validationData.Errors, ValidationError{
		Field: key,
		Text:  err,
	})
}

func (fv *formValidation) Redirect(url string) {
	fv.validationData.RedirectionLocation = url
}

func (fv *formValidation) AfterContent(content template.HTML) {
	fv.validationData.AfterContent = content
}

func (fv *formValidation) Valid() bool {
	return len(fv.validationData.Errors) == 0
}
