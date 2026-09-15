package prago

import (
	"github.com/hypertornado/prago/pragoai"
)

func (app *App) initAIClient() {
	app.AI = pragoai.NewClaudeClient(app.mustGetSetting("claude_api_key"))
}

func (app *App) initAI() {

	app.NewSetting("claude_api_key", "sysadmin").ValueChangeCallback(func() {
		app.initAIClient()
	})

	app.initAIClient()

	app.ActionForm("_aichat", func(form *Form, request *Request) {

		var modelOptions [][2]string
		models, err := app.AI.GetModels()
		must(err)
		for _, model := range models {
			modelOptions = append(modelOptions, [2]string{
				model.ID,
				model.Name,
			})
		}

		form.AddSelect("model", "Modely", modelOptions).SetValue(pragoai.DefaultModelLow)

		form.AddTextarea("text", "Text").SetFocused()
		form.AddSubmit("Odeslat")
	}, func(fv FormValidation, request *Request) {

		text := request.Param("text")
		if text == "" {
			fv.AddItemError("text", "Zadejte text")
		}

		if !fv.Valid() {
			return
		}

		response, err := app.AI.GetText(pragoai.Request{
			Text:  text,
			Model: request.Param("model"),
		})

		if err != nil {
			fv.AddError(err.Error())
			return
		}

		responseHTML, err := markdownToHTML(response)
		if err != nil {
			fv.AddError(err.Error())
			return
		}

		fv.AfterContent(responseHTML)

	}).Permission("sysadmin").Name(unlocalized("AI")).Board(app.optionsBoard).Icon(iconAI)
}
