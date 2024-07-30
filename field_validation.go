package prago

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func (app *App) initFieldValidations() {
	for _, resource := range app.resources {
		for _, field := range resource.fields {
			validations := field.tags["prago-validations"]
			if validations != "" {
				for _, v := range strings.Split(validations, ",") {
					field.addPragoFieldValidation(v)
				}
			}
		}
	}
}

func (field *Field) addPragoFieldValidation(nameOfValidation string) {
	if nameOfValidation == "nonempty" {
		field.addValidationNonempty()
		return
	}

	if nameOfValidation == "strong" {
		field.addValidationStrongRelation()
		return
	}

	panic(
		fmt.Sprintf("can't add validation on field '%s' of resource '%s': unknown validation name: %s", field.id, field.resource.pluralName("en"), nameOfValidation))

}

func (field *Field) addValidationStrongRelation() {

	if !field.fieldType.IsRelation() {
		panic(
			fmt.Sprintf("field %s (resource %s) is not of type relation", field.id, field.resource.id),
		)
	}

	field.relatedResource.addDeleteValidation(func(item any, v Validation, ud UserData) {
		itemsVal := reflect.ValueOf(item).Elem()
		fieldID := itemsVal.FieldByName("ID").Int()
		relatedCount, err := field.resource.query(context.Background()).Is(field.id, fieldID).count()
		must(err)
		if relatedCount > 0 {
			v.AddError(
				"Can't delete item with strong relation",
			)
		}
	})

}

func (field *Field) addValidationNonempty() {
	if field.tags["prago-required"] != "false" {
		field.required = true
	}

	field.Validation(func(fieldVal any, userData UserData) error {
		typ := reflect.TypeOf(fieldVal)
		valid := true
		if typ.Kind() == reflect.Int64 ||
			typ.Kind() == reflect.Int32 ||
			typ.Kind() == reflect.Int {

			intVal := fieldVal.(int64)
			if intVal == 0 {
				valid = false
			}

		}
		if typ.Kind() == reflect.Float64 ||
			typ.Kind() == reflect.Float32 {

			floatVal := fieldVal.(float64)
			if floatVal == 0 {
				valid = false
			}
		}

		if typ.Kind() == reflect.String {
			if fieldVal.(string) == "" {
				valid = false
			}
		}

		if !valid {
			return errors.New(messages.Get(userData.Locale(), "admin_validation_not_empty"))
		}
		return nil
	})

}

func (field *Field) Validation(fn func(fieldVal any, userData UserData) error) {

	field.resource.addUpdateValidation(func(item any, v Validation, ud UserData) {

		itemsVal := reflect.ValueOf(item).Elem()
		fieldVal := itemsVal.FieldByName(field.fieldClassName)

		err := fn(fieldVal.Interface(), ud)
		if err != nil {
			v.AddItemError(field.id, err.Error())
		}

	})

}
