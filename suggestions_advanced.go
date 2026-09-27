package prago

import (
	"fmt"
	"net/url"
	"strconv"
)

func (app *App) initSuggestionsAdvanced() {

	app.initSuggestionsAdvancedSelect()

	app.PopupForm("_suggestionsadvanced", func(form *Form, request *Request) {

		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		form.AddHidden("_resource").SetValue(resource.id)

		var orderOptions [][2]string
		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			if !field.canOrder {
				continue
			}

			orderOptions = append(orderOptions, [2]string{
				field.id,
				field.GetName("cs"),
			})
			orderOptions = append(orderOptions, [2]string{
				field.id + "_desc",
				field.GetName("cs") + " (sestupně)",
			})
		}

		orderItem := form.AddSelect("_order", "Seřadit", orderOptions).SetTextOver("")
		if resource.orderDesc {
			orderItem.SetValue(resource.orderByColumn + "_desc")
		} else {
			orderItem.SetValue(resource.orderByColumn)
		}

		form.AddNumber("_count", "Počet výsledků").SetValue("100").SetTextOver("")

		var autofocused bool

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}

			var autofocus = false
			if !autofocused && field.id != "id" {
				autofocus = true
				autofocused = true
			}

			listFilterFormAdd(form, field, "", request, autofocus)
		}

		form.AddSubmit("Najít")

	}, func(fv FormValidation, request *Request) {

		resource := app.getResourceByID(request.Param("_resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		var values url.Values = map[string][]string{}
		values.Set("_resource", resource.id)

		count, err := strconv.Atoi(request.Param("_count"))
		if err != nil || count <= 0 {
			fv.AddItemError("_count", "Neplatné číslo")
		} else {
			values.Set("_count", fmt.Sprintf("%d", count))
		}
		values.Set("_order", request.Param("_order"))

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}

			val := listFilterFormGetValues(fv, field, request)
			if val != "" {
				values.Set(field.id, val)
			}
		}

		if !fv.Valid() {
			return
		}

		fv.PopupFormURL(fmt.Sprintf("/admin/_suggestionsadvancedselect?%s", values.Encode()))

	}).Permission(loggedPermission).Name(Unlocalized("Filtr")).Icon("glyphicons-basic-395-filter.svg")

}
