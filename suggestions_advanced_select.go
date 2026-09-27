package prago

import (
	"reflect"
	"strconv"
)

func (app *App) initSuggestionsAdvancedSelect() {

	app.PopupForm("_suggestionsadvancedselect", func(form *Form, request *Request) {

		resource := app.getResourceByID(request.Param("_resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		form.AddHidden("_resource").SetValue(resource.id)

		count, err := strconv.Atoi(request.Param("_count"))
		Must(err)
		if count <= 0 || count > 10000 {
			panic("wrong count number")
		}

		var filter = map[string]string{}
		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			val := request.Param(field.id)
			if val != "" {
				filter[field.id] = val
			}
		}

		q := resource.query(request.r.Context())

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}

			if request.Param("_order") == field.id {
				q.Order(field.id)
			}

			if request.Param("_order") == field.id+"_desc" {
				q.OrderDesc(field.id)
			}
		}

		q.Limit(int64(count))
		q = resource.addFilterToQuery(q, filter)
		items, err := q.list()
		Must(err)

		itemsVal := reflect.ValueOf(items)

		if itemsVal.Len() > 0 {
			var options []*FormOption
			for i := 0; i < itemsVal.Len(); i++ {
				item := itemsVal.Index(i).Interface()
				previewer := resource.previewer(request, item)
				options = append(options, previewerToFormOption(previewer))
			}

			form.AutosubmitOnDataChange = true

			form.AddRadioOptions("item", "", options)
		} else {
			form.AddText("_", "").SetValue("Nic nenalezeno").SetReadonly()
		}

	}, func(fv FormValidation, request *Request) {

		resource := app.getResourceByID(request.Param("_resource"))
		if !request.Authorize(resource.canView) {
			panic("not allowed")
		}

		id := request.Param("item")
		if id == "" {
			fv.AddItemError("item", "Vyberte položku")
		}

		if !fv.Valid() {
			return
		}

		item := resource.query(request.r.Context()).ID(id)
		preview := resource.previewer(request, item).Preview(nil)
		fv.Data(previewtoSuggestion(preview))

	}).Permission(loggedPermission).Name(Unlocalized("Vyberte")).Icon("glyphicons-basic-395-filter.svg")

}
