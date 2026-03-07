package prago

func (app *App) initAI() {

	ActionForm(app, "_aichat", func(form *Form, request *Request) {
		form.AddTextareaInput("text", "Text").Focused = true
		form.AddSubmit("Odeslat")
	}, func(fv FormValidation, request *Request) {

	}).Permission("sysadmin").Name(unlocalized("AI")).Board(app.optionsBoard)

}
