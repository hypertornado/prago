package prago

import (
	"fmt"
	"time"
)

func (app *App) initAIUsage() {

	app.ActionForm("_aiusage", func(form *Form, request *Request) {

		form.AddDateTimePicker("from", "Od").SetValue(time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04"))
		form.AddDateTimePicker("to", "Do").SetValue(time.Now().Format("2006-01-02 15:04"))

		form.AddSubmit("Zobrazit")
	}, func(fv FormValidation, request *Request) {

		fmt.Println(request.Param("from"))

		from, err := time.Parse("2006-01-02T15:04", request.Param("from"))
		if err != nil {
			fv.AddItemError("from", "Chyba")
		}

		to, err := time.Parse("2006-01-02T15:04", request.Param("to"))
		if err != nil {
			fv.AddItemError("to", "Chyba")
		}

		if !fv.Valid() {
			return
		}

		usages, err := app.AI.GetUsage(from, to)
		if err != nil {
			fv.AddError(err.Error())
		}

		table := app.NewTable()
		table.Header("Model", "UncachedInputTokens", "CacheReadInputTokens", "CacheCreationInputTokens", "OutputTokens")
		for _, usage := range usages {
			table.Row(
				table.Cell(usage.Model),
				table.Cell(usage.UncachedInputTokens),
				table.Cell(usage.CacheReadInputTokens),
				table.Cell(usage.CacheCreationInputTokens),
				table.Cell(usage.OutputTokens),
			)
		}
		fv.AfterContent(table.ExecuteHTML())

	}).Permission("sysadmin").Name(Unlocalized("AI Usage")).Board(aiBoard)

}
