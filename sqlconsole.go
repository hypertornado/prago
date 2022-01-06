package prago

import (
	"fmt"
)

func (app *App) initSQLConsole() {

	app.FormAction("sqlconsole").Name(unlocalized("SQL Console")).Permission(sysadminPermission).IsWide().Form(
		func(form *Form, request *Request) {
			form.Title = "SQL Console"
			form.AddTextareaInput("q", "").Focused = true
			form.AddSubmit("Execute SQL")
		},
	).Validation(func(vc ValidationContext) {
		q := vc.GetValue("q")
		var message string
		table := app.Table()

		if q != "" {
			rows, err := app.db.QueryContext(vc.Request().Request().Context(), q)
			rowCount := 0
			if err != nil {
				message = err.Error()
			} else {
				columns, err := rows.Columns()
				must(err)
				var header []interface{}
				for _, v := range columns {
					header = append(header, v)
				}
				table.Header(header...)

				count := len(columns)
				values := make([]interface{}, count)
				valuePtrs := make([]interface{}, count)

				for rows.Next() {
					rowCount += 1
					for i := range columns {
						valuePtrs[i] = &values[i]
					}

					rows.Scan(valuePtrs...)

					var row []interface{}
					for i := range columns {
						val := values[i]

						b, ok := val.([]byte)
						var v interface{}
						if ok {
							v = string(b)
						} else {
							v = val
						}

						row = append(row,
							fmt.Sprintf("%v", v),
						)
					}
					table.Row(row...)
				}
			}
			table.AddFooterText(fmt.Sprintf("%d items", rowCount))
		}

		if message != "" {
			vc.AddError(message)
		}

		vc.Validation().AfterContent = table.ExecuteHTML()
	})

}
