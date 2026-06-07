package prago

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/url"
)

func (app *App) initSQLView() {

	ActionUI(app, "_sqlview", func(request *Request) template.HTML {

		table := app.Table()

		tablesArr := listTablesArr(app)

		var resourcesTableNames = map[string]*Resource{}
		for _, v := range app.resources {
			resourcesTableNames[v.id] = v
		}

		for _, tableName := range tablesArr {

			var showDelete = true
			var useIcon = "🔴"
			if resourcesTableNames[tableName] != nil {
				useIcon = "✅"
				showDelete = false
			}

			lastCell := Cell("")
			if showDelete {
				q := fmt.Sprintf("DROP TABLE %s;", tableName)
				lastCell.Button(&TableCellButton{
					Name:    "Delete table",
					OnClick: template.JS(fmt.Sprintf("popup(\"/admin/_sqlconsole?q=%s\")", url.QueryEscape(q))),
				})
			}

			tableSize := app.getTableDataSize(tableName)

			table.Row(Cell(useIcon), Cell(tableName).Header(), Cell(fmt.Sprintf("%d B", tableSize)), lastCell)

			rows, err := app.db.Query(fmt.Sprintf("DESCRIBE %s;", tableName))
			if err != nil {
				panic(err)
			}
			defer rows.Close()

			var field, Type, Null, Key, Default, Extra sql.NullString

			resource := resourcesTableNames[tableName]

			for rows.Next() {
				err = rows.Scan(&field, &Type, &Null, &Key, &Default, &Extra)
				if err != nil {
					panic(err)
				}

				fieldName := field.String

				columnCell := Cell(fieldName)

				var canDeleteField bool = true
				if resource != nil && resource.fieldMap[field.String] != nil {
					canDeleteField = false
				}

				q := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", tableName, fieldName)

				deleteFieldCell := Cell("")
				if canDeleteField {
					deleteFieldCell.Button(&TableCellButton{
						Name:    "Delete field",
						OnClick: template.JS(fmt.Sprintf("popup(\"/admin/_sqlconsole?q=%s\")", url.QueryEscape(q))),
					})
				}

				indexQ := fmt.Sprintf("CREATE INDEX `idx_%s` ON `%s` (`%s`);", fieldName, tableName, fieldName)
				deleteFieldCell.Button(&TableCellButton{
					Name:    "Create index",
					OnClick: template.JS(fmt.Sprintf("popup(\"/admin/_sqlconsole?q=%s\")", url.QueryEscape(indexQ))),
				})

				for _, indexName := range app.getSQLIndexNames(tableName, fieldName) {
					columnCell.DescriptionAfter("Index: " + indexName)
				}

				table.Row(Cell(""), columnCell, Cell(fmt.Sprintf("Type: %-15s Null: %-5s Key: %-5s Default: %-10s Extra: %s\n",
					Type.String, Null.String, Key.String, Default.String, Extra.String)), deleteFieldCell)

			}
		}
		return table.ExecuteHTML()
	}).Permission("sysadmin").Board(sysadminBoard).Name(unlocalized("SQL View"))

}

func (app *App) getSQLIndexNames(tableName, fieldName string) (ret []string) {

	q := fmt.Sprintf(`
SELECT INDEX_NAME 
FROM information_schema.STATISTICS 
WHERE TABLE_SCHEMA = '%s' 
  AND TABLE_NAME = '%s'
  AND COLUMN_NAME = '%s';`, app.dbConfig.Name, tableName, fieldName)

	rows, err := app.db.Query(q)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var rowName sql.NullString
	for rows.Next() {
		err = rows.Scan(&rowName)
		if err != nil {
			panic(err)
		}
		ret = append(ret, rowName.String)
	}
	return
}
