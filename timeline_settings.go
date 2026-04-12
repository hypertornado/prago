package prago

func (app *App) initTimelineSettings() {

	PopupForm(app, "_timeline-settings", func(form *Form, request *Request) {

		timeline, err := app.getTimeline(request, request.Param("_uuid"))
		must(err)

		typeValues := [][2]string{
			{"day", "Den"},
			{"month", "Měsíc"},
			{"year", "Rok"},
		}
		typeItem := form.AddRadio("_type", "Typ zobrazení", typeValues)
		typeItem.Value = request.Param("_type")
		typeItem.Icon = "glyphicons-basic-46-calendar.svg"

		alignValues := [][2]string{
			{"history", "Do minulosti"},
			{"center", "Na střed"},
			{"future", "Do budoucnosti"},
		}
		alignItem := form.AddRadio("_alignment", "Zarovnání časové osy", alignValues)
		alignItem.Value = request.Param("_alignment")
		alignItem.Icon = "glyphicons-basic-749-resize-horizontal.svg"

		if timeline.optionsForm != nil {
			timeline.optionsForm(request, form)
		}

		form.AddSubmit("Nastavit")
	}, func(fv FormValidation, request *Request) {
		data := map[string]string{}
		params := request.Params()
		for k := range params {
			data[k] = params.Get(k)
		}
		fv.Data(data)
	}).Permission(loggedPermission).Name(unlocalized("Nastavení časové osy")).Icon("glyphicons-basic-137-cogwheel.svg")

}
