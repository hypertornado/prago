package prago

import (
	"bytes"
	"fmt"
)

type SQLConsoleCell struct {
	Value string
}

func (app *App) initSQLConsole() {
	app.Action("sqlconsole").Name(unlocalized("SQL Console")).Permission(sysadminPermission).Template("admin_form").IsWide().DataSource(
		func(request *Request) interface{} {
			form := NewForm("/admin/api/sqlconsole")
			form.Title = "SQL Console"
			form.AddTextareaInput("q", "").Focused = true
			form.AddHidden("csrf").Value = request.csrfToken()
			form.AddSubmit("_submit", "Execute SQL")
			return form
		},
	)

	app.API("sqlconsole").Method("POST").Permission(sysadminPermission).HandlerJSON(func(request *Request) interface{} {
		validation := NewFormValidation()

		q := request.Params().Get("q")
		var message string
		var table [][]SQLConsoleCell

		if q != "" && request.csrfToken() != request.Params().Get("csrf") {
			panic("wrong csrf")
		}

		if q != "" {
			rows, err := app.db.QueryContext(request.Request().Context(), q)
			if err != nil {
				message = err.Error()
			} else {
				columns, err := rows.Columns()
				must(err)
				var header []SQLConsoleCell
				for _, v := range columns {
					header = append(header, SQLConsoleCell{
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

					var row []SQLConsoleCell
					for i := range columns {
						val := values[i]

						b, ok := val.([]byte)
						var v interface{}
						if ok {
							v = string(b)
						} else {
							v = val
						}

						row = append(row, SQLConsoleCell{
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
			validation.AddError(message)
		}

		bufStats := new(bytes.Buffer)
		err := request.app.ExecuteTemplate(bufStats, "sql_console", retData)
		if err != nil {
			panic(err)
		}
		validation.AfterContent = bufStats.String()
		return validation
	})

}
