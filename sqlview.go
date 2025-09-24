package prago

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/url"
)

func (app *App) initSQLView() {

	ActionUI(app, "_sqlview", func(r *Request) template.HTML {

		table := app.Table()

		tablesArr := listTablesArr(app)

		var resourcesTableNames = map[string]*Resource{}
		for _, v := range app.resources {
			resourcesTableNames[v.id] = v
		}

		for _, v := range tablesArr {

			var showDelete = true
			var useIcon = "🔴"
			if resourcesTableNames[v] != nil {
				useIcon = "✅"
				showDelete = false
			}

			lastCell := Cell("")
			if showDelete {
				q := fmt.Sprintf("DROP TABLE %s;", v)
				lastCell.Button(&TableCellButton{
					Name: "Delete table",
					URL:  "/admin/_sqlconsole?q=" + url.QueryEscape(q),
				})
			}

			tableSize := app.getTableDataSize(v)

			table.Row(Cell(useIcon), Cell(v).Header(), Cell(fmt.Sprintf("%d B", tableSize)), lastCell)

			rows, err := app.db.Query(fmt.Sprintf("DESCRIBE %s;", v))
			if err != nil {
				panic(err)
			}
			defer rows.Close()

			var field, Type, Null, Key, Default, Extra sql.NullString

			//var hasFields = map[string]bool{}
			resource := resourcesTableNames[v]

			//resourcesTableNames[]

			for rows.Next() {
				err = rows.Scan(&field, &Type, &Null, &Key, &Default, &Extra)
				if err != nil {
					panic(err)
				}

				columnCell := Cell(field.String)

				var canDeleteField bool = true
				if resource != nil && resource.fieldMap[field.String] != nil {
					canDeleteField = false
				}

				q := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", v, field.String)

				deleteFieldCell := Cell("")
				if canDeleteField {
					deleteFieldCell.Button(&TableCellButton{
						Name: "Delete field",
						URL:  "/admin/_sqlconsole?q=" + url.QueryEscape(q),
					})
				}

				table.Row(Cell(""), columnCell, Cell(fmt.Sprintf("Type: %-15s Null: %-5s Key: %-5s Default: %-10s Extra: %s\n",
					Type.String, Null.String, Key.String, Default.String, Extra.String)), deleteFieldCell)

			}
		}
		return table.ExecuteHTML()
	}).Permission("sysadmin").Board(sysadminBoard).Name(unlocalized("SQL View"))

}
