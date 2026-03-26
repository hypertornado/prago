package prago

import (
	"context"
	"fmt"
	"reflect"
	"sort"
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
		form.AddHidden("resource").Value = request.Param("resource")
		form.AutosubmitFirstTime = true

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canUpdate) {
			fv.AddError("Not allowed")
			return
		}
		fv.RunTask(request, func(fta *FormTaskActivity) error {
			fta.TableCells(Cell("Item").Header(), Cell("Validation errors").Header())
			return resource.consistencyCheck(request, fta)
		})

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

			form.AddSubmit("Validate resource")
		}, func(vc FormValidation, request *Request) {

			resource := app.getResourceByID(request.Param("resource"))
			if resource == nil {
				vc.AddItemError("resource", "Select resource")
			}

			if !vc.Valid() {
				return
			}

			vc.RunTask(request, func(fta *FormTaskActivity) error {
				fta.TableCells(Cell("Item").Header(), Cell("Validation errors").Header())
				return resource.consistencyCheck(request, fta)
			})

		}).Name(unlocalized("Validate resources")).Permission(sysadminPermission).Board(sysadminBoard)
}

func (resource *Resource) consistencyCheck(request *Request, fta *FormTaskActivity) error {
	var correctCount, incorrectCount int64

	var i int64 = -1

	allItems := resource.countAllItems()

	resource.forEach(context.Background(), func(a any) error {

		i++

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

		fta.Progress(i, allItems)
		fta.Description(fmt.Sprintf("Checking #%d", id))

		if validationResult != "" {
			fta.TableCells(
				Cell(fmt.Sprintf("#%d", id)).URL(fmt.Sprintf("/admin/%s/%d", resource.id, id)),
				Cell(validationResult),
			)
		}

		//time.Sleep(10 * time.Millisecond)

		return nil
	})

	fta.Description(fmt.Sprintf("Checking done, correct: %d, incorrect: %d", correctCount, incorrectCount))

	return nil

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
