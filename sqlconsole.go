package prago

import (
	"fmt"
)

func (app *App) initSQLConsole() {

	ActionForm(app, "sqlconsole",
		func(form *Form, request *Request) {
			form.Title = "SQL Console"
			form.AddTextareaInput("q", "").Focused = true
			form.AddSubmit("Execute SQL")
		}, func(vc FormValidation, request *Request) {
			q := request.Param("q")
			var message string
			table := app.Table()

			if q != "" {
				rows, err := app.db.QueryContext(request.Request().Context(), q)
				rowCount := 0
				if err != nil {
					message = err.Error()
				} else {
					columns, err := rows.Columns()
					must(err)
					table.Header(columns...)

					count := len(columns)
					values := make([]interface{}, count)
					valuePtrs := make([]interface{}, count)

					var cells []*TableCell

					for rows.Next() {
						rowCount += 1
						for i := range columns {
							valuePtrs[i] = &values[i]
						}

						rows.Scan(valuePtrs...)

						for i := range columns {
							val := values[i]

							b, ok := val.([]byte)
							var v interface{}
							if ok {
								v = string(b)
							} else {
								v = val
							}

							cells = append(cells, Cell(fmt.Sprintf("%v", v)))

						}
						table.Row(cells...)
						cells = nil
					}
				}
				table.AddFooterText(fmt.Sprintf("%d items", rowCount))
			}

			if message != "" {
				vc.AddError(message)
			}

			vc.AfterContent(table.ExecuteHTML())
		}).Name(unlocalized("SQL Console")).Permission(sysadminPermission).Board(sysadminBoard)
}
