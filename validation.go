package prago

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Validation interface {
	AddError(err string)
	AddItemError(key, err string)
	Valid() bool
}

type ValidationError struct {
	OK    bool
	Field string
	Text  string
}

func (app *App) initValidation() {

	PopupForm(app, "_validation-consistency", func(form *Form, request *Request) {

		form.AddTextInput("offset", "Offset")
		form.AddTextInput("limit", "Limit")

		form.AddHidden("resource").Value = request.Param("resource")
		form.AddSubmit("Spustit kontrolu konzistence")
		//form.AutosubmitFirstTime = true

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canUpdate) {
			fv.AddError("Not allowed")
			return
		}

		var offset, limit int
		var err error
		if request.Param("offset") != "" {
			offset, err = strconv.Atoi(request.Param("offset"))
			if err != nil || offset < 0 {
				fv.AddItemError("offset", "Špatná hodnota")
			}
		}
		if request.Param("limit") != "" {
			limit, err = strconv.Atoi(request.Param("limit"))
			if err != nil || limit < 0 {
				fv.AddItemError("limit", "Špatná hodnota")
			}
		}

		table := resource.validateResourceTable(request, int64(offset), int64(limit), true)
		fv.AfterContent(table.ExecuteHTML())
	}).Permission(loggedPermission).Name(unlocalized("Kontrola konzistence"))

	ActionForm(app, "_validations",
		func(form *Form, request *Request) {
			form.Title = "Validations check"

			var values [][2]string
			values = append(values, [2]string{})
			for _, v := range app.resources {
				values = append(values, [2]string{
					v.id,
					v.id + " - " + v.pluralName("en"),
				})
			}

			sort.Slice(values, func(i, j int) bool {
				return strings.Compare(values[i][0], values[j][0]) < 0
			})

			form.AddSelect("resource", "Resource", values)
			form.AddCheckbox("ignorecorrect", "Don't show correct items").Value = "on"

			form.AddSubmit("Validate resource")
		}, func(vc FormValidation, request *Request) {

			resource := app.getResourceByID(request.Param("resource"))
			if resource == nil {
				vc.AddItemError("resource", "Select resource")
			}

			if !vc.Valid() {
				return
			}

			var ignorecorrect bool
			if request.Param("ignorecorrect") == "on" {
				ignorecorrect = true
			}

			table := resource.validateResourceTable(request, 0, 0, ignorecorrect)

			vc.AfterContent(table.ExecuteHTML())

		}).Name(unlocalized("Validate resources")).Permission(sysadminPermission).Board(sysadminBoard)
}

func (resource *Resource) validateResourceTable(request *Request, offset, limit int64, ignorecorrect bool) *Table {
	table := resource.app.Table()

	table.Header("Item", "Validation errors")
	var correctCount, incorrectCount int64

	var i int64 = -1

	resource.forEach(request.r.Context(), func(a any) error {

		i++

		if i < offset {
			return nil
		}
		if limit > 0 && i >= offset+limit {
			return nil
		}

		val := reflect.ValueOf(a).Elem()
		id := val.FieldByName("ID").Int()

		var validationResult string
		result := resource.validateUpdate(a, request)
		validationResult = result.TextErrorReport(id, "en").Text

		if validationResult == "" {
			correctCount++
		} else {
			incorrectCount++
		}

		if !ignorecorrect || validationResult != "" {
			table.Row(
				Cell(fmt.Sprintf("#%d", id)).URL(fmt.Sprintf("/admin/%s/%d", resource.id, id)),
				Cell(validationResult),
			)
		}
		return nil
	})

	table.AddFooterText(fmt.Sprintf("correct: %d, incorrect: %d", correctCount, incorrectCount))

	return table
}

func (resource *Resource) validateUpdate(item any, user UserData) *itemValidation {
	itemValidation := newItemValidation()
	for _, validation := range resource.updateValidations {
		validation(item, itemValidation, user)
	}
	return itemValidation
}

func (resource *Resource) validateDelete(item any, user UserData) *itemValidation {
	itemValidation := newItemValidation()
	for _, validation := range resource.deleteValidations {
		validation(item, itemValidation, user)
	}
	return itemValidation
}

func TestValidationUpdate[T any](app *App, item *T, user UserData) ([]ValidationError, bool) {
	resource := getResource[T](app)
	validation := resource.validateUpdate(item, user)
	return validation.errors, validation.Valid()
}

func TestValidationDelete[T any](app *App, item *T, user UserData) ([]ValidationError, bool) {
	resource := getResource[T](app)
	validation := resource.validateDelete(item, user)
	return validation.errors, validation.Valid()
}
