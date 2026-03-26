package prago

import (
	"strconv"
	"strings"
)

func (app *App) initListSettings() {

	PopupForm(app, "_list-items-per-page", func(form *Form, request *Request) {
		form.AddHidden("resource").Value = request.Param("resource")

		item := form.AddSelect("count", "Počet položek na stránce", getStatsLimitSelectPlain())
		item.Value = request.Param("count")
		item.Focused = true

		form.AddSubmit("Nastavit počet")

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canUpdate) {
			fv.AddError("Not allowed")
			return
		}
		count, err := strconv.Atoi(request.Param("count"))
		if err != nil || count <= 0 {
			fv.AddError("Špatný počet")
			return
		}
		fv.Data(count)
	}).Permission(loggedPermission).Icon("glyphicons-basic-960-files-queue.svg").Name(unlocalized("Počet položek na stránce"))

	PopupForm(app, "_list-items-visible", func(form *Form, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canUpdate) {
			panic("can't show")
		}

		fieldsMap := map[string]bool{}
		fields := strings.Split(request.Param("fields"), ",")
		for _, field := range fields {
			fieldsMap[field] = true
		}

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			input := form.AddCheckbox(field.id, field.name(request.Locale()))
			if fieldsMap[field.id] {
				input.Value = "on"
			}
		}

		form.AddHidden("resource").Value = request.Param("resource")
		form.AddSubmit("Nastavit viditelné sloupce")

	}, func(fv FormValidation, request *Request) {
		resource := app.getResourceByID(request.Param("resource"))
		if !request.Authorize(resource.canView) {
			fv.AddError("Not allowed")
			return
		}

		var ret []string

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			if request.Param(field.id) != "on" {
				continue
			}
			ret = append(ret, field.id)
		}

		fv.Data(strings.Join(ret, ","))
	}).Permission(loggedPermission).Icon("glyphicons-basic-107-text-width.svg").Name(unlocalized("Viditelné sloupce"))

}
