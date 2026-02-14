package prago

import (
	"html/template"
)

type FormValidation interface {
	AddError(err string)
	AddOK(err string)
	AddItemError(key, err string)
	Valid() bool
	Redirect(string)
	AfterContent(template.HTML)
	Data(any)
}

type formValidationData struct {
	RedirectionLocation string
	AfterContent        template.HTML
	Errors              []ValidationError
	Data                any
}

type formValidationReport struct {
	Text string
}

func newFormValidationData() *formValidationData {
	ret := &formValidationData{}
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
	fv.validationData.Errors = append(fv.validationData.Errors, ValidationError{
		Text: err,
	})
}

func (fv *formValidation) AddOK(message string) {
	fv.validationData.Errors = append(fv.validationData.Errors, ValidationError{
		OK:   true,
		Text: message,
	})
}

func (fv *formValidation) AddItemError(key, err string) {
	fv.validationData.Errors = append(fv.validationData.Errors, ValidationError{
		Field: key,
		Text:  err,
	})
}

func (fv *formValidation) Redirect(url string) {
	fv.validationData.RedirectionLocation = url
}

func (fv *formValidation) Data(data any) {
	fv.validationData.Data = data

}

func (fv *formValidation) AfterContent(content template.HTML) {
	fv.validationData.AfterContent = content
}

func (fv *formValidation) Valid() bool {
	return len(fv.validationData.Errors) == 0
}
