package prago

import "fmt"

type SQLConsoleCell struct {
	Value string
}

func (app *App) initSQLConsole() {
	app.Action("sqlconsole").Name(unlocalized("SQL Console")).Permission(sysadminPermission).Template("sql_console").IsWide().DataSource(
		func(request *Request) interface{} {
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

			return map[string]interface{}{
				"csrf":    request.csrfToken(),
				"message": message,
				"q":       q,
				"table":   table,
			}
		},
	)
}
