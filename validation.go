package prago

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

var errValidation = errors.New("validation error")

type formValidation struct {
	Valid               bool
	RedirectionLocation string
	AfterContent        template.HTML
	Errors              []FormValidationError
	ItemErrors          map[string][]FormValidationError
}

type FormValidationError struct {
	Text string
}

type formValidationReport struct {
	Text string
}

func (validation *formValidation) AddError(text string) {
	validation.Valid = false
	validation.Errors = append(validation.Errors, FormValidationError{
		Text: text,
	})
}

func (validation *formValidation) AddItemError(id, text string) {
	validation.Valid = false
	validation.ItemErrors[id] = append(validation.ItemErrors[id], FormValidationError{
		Text: text,
	})
}

func (validation *formValidation) TextErrorReport(id int64, locale string) formValidationReport {
	var errors []string
	for _, v := range validation.Errors {
		errors = append(errors, v.Text)
	}

	for k, v := range validation.ItemErrors {
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

func newFormValidation() *formValidation {
	ret := &formValidation{
		Valid: true,
	}
	ret.ItemErrors = map[string][]FormValidationError{}
	return ret
}

type Validation interface {
	Locale() string
	GetValue(key string) string
	GetValues(key string) []string
	AddError(err string)
	AddItemError(key, err string)
	Validation() *formValidation
	Valid() bool
	//Request() *Request
	UserHasPermission(Permission) bool
	Context() context.Context
}

type requestValidation struct {
	request    *Request
	validation *formValidation
}

func newRequestValidation(request *Request) *requestValidation {
	return &requestValidation{
		request:    request,
		validation: newFormValidation(),
	}
}

func (rv *requestValidation) Locale() string {
	return rv.request.Locale()
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

func (rv *requestValidation) Validation() *formValidation {
	return rv.validation
}

func (rv *requestValidation) Valid() bool {
	return rv.validation.Valid
}

/*func (rv *requestValidation) Request() *Request {
	return rv.request
}*/

func (rv *requestValidation) Context() context.Context {
	return rv.request.r.Context()
}

func (rv *requestValidation) UserHasPermission(permission Permission) bool {
	return rv.request.Authorize(permission)
}

type valuesValidation struct {
	values     url.Values
	validation *formValidation
	request    UserData
	app        *App
	ctx        context.Context
}

func newValuesValidation(ctx context.Context, app *App, request UserData, values url.Values) *valuesValidation {
	return &valuesValidation{
		values:     values,
		validation: newFormValidation(),
		request:    request,
		app:        app,
		ctx:        ctx,
	}
}

func (rv *valuesValidation) Locale() string {
	return rv.request.Locale()
}

func (rv *valuesValidation) GetValue(key string) string {
	return rv.values.Get(key)
}

func (rv *valuesValidation) GetValues(key string) []string {
	return rv.values[key]
}

func (rv *valuesValidation) AddError(err string) {
	rv.validation.AddError(err)
}

func (rv *valuesValidation) AddItemError(key, err string) {
	rv.validation.AddItemError(key, err)
}

func (rv *valuesValidation) Validation() *formValidation {
	return rv.validation
}

func (rv *valuesValidation) Valid() bool {
	return rv.validation.Valid
}

func (rv *valuesValidation) Request() *Request {
	return nil
}

func (rv *valuesValidation) UserHasPermission(permission Permission) bool {
	return rv.request.Authorize(permission)
}

func (rv *valuesValidation) Context() context.Context {
	return rv.ctx
}
