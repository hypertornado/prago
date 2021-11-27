package prago

import (
	"bytes"
	"fmt"
)

type sqlConsoleCell struct {
	Value string
}

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
		var table [][]sqlConsoleCell

		if q != "" {
			rows, err := app.db.QueryContext(vc.Request().Request().Context(), q)
			if err != nil {
				message = err.Error()
			} else {
				columns, err := rows.Columns()
				must(err)
				var header []sqlConsoleCell
				for _, v := range columns {
					header = append(header, sqlConsoleCell{
						Value: v,
					})
				}
				table = append(table, header)

				count := len(columns)
				values := make([]interface{}, count)
				valuePtrs := make([]interface{}, count)

				for rows.Next() {
					for i := range columns {
						valuePtrs[i] = &values[i]
					}

					rows.Scan(valuePtrs...)

					var row []sqlConsoleCell
					for i := range columns {
						val := values[i]

						b, ok := val.([]byte)
						var v interface{}
						if ok {
							v = string(b)
						} else {
							v = val
						}

						row = append(row, sqlConsoleCell{
							Value: fmt.Sprintf("%v", v),
						},
						)

					}
					table = append(table, row)
				}
			}
		}

		retData := map[string]interface{}{
			"table": table,
		}

		if message != "" {
			vc.AddError(message)
		}

		bufStats := new(bytes.Buffer)
		err := app.ExecuteTemplate(bufStats, "sql_console", retData)
		if err != nil {
			panic(err)
		}
		vc.Validation().AfterContent = bufStats.String()
	})

}
