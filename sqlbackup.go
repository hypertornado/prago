package prago

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func (app *App) initSQLBackup() {

	var backupFilePath string

	app.addCommand("sqlbackup").Description("Backup SQL to FILE").flag(
		newCommandFlag("path", "path of sql backup file").String(&backupFilePath),
	).Callback(func() {
		if backupFilePath == "" {
			fmt.Println("No backupFilePath set")
			return
		}

		file, err := os.Create(backupFilePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		err = app.backupSQL(file, nil)
		if err != nil {
			fmt.Println("Error backing SQL file:", err)
			return
		}
	})

	var restoreFilePath string
	app.addCommand("sqlrestore").Description("Restore SQL from FILE").flag(
		newCommandFlag("path", "path of sql backup file").String(&restoreFilePath),
	).Callback(func() {
		if restoreFilePath == "" {
			fmt.Println("No restoreFilePath set")
			return
		}

		file, err := os.Open(restoreFilePath)
		if err != nil {
			fmt.Printf("Error opening file '%s': %s\n", restoreFilePath, err)
			return
		}
		defer file.Close()

		err = app.restoreSQLBackup(file)
		if err != nil {
			fmt.Println("Error restoring SQL file:", err)
			return
		}
	})

	ActionForm(app, "_sqlbackup",
		func(form *Form, request *Request) {
			form.Title = "SQL Backup"

			/*tables, err := listTables(app.db)
			must(err)

			var tablesAr []string

			for k := range tables {
				tablesAr = append(tablesAr, k)
			}

			sort.Strings(tablesAr)*/

			tablesArr := listTablesArr(app)

			var resourcesTableNames = map[string]bool{}
			for _, v := range app.resources {
				resourcesTableNames[v.id] = true
			}

			for _, v := range tablesArr {
				tableSize := app.getTableDataSize(v)

				var defaultIgnore = true
				var useIcon = "🔴"
				if resourcesTableNames[v] {
					useIcon = "✅"
					defaultIgnore = false
				}

				item := form.AddCheckbox(v, fmt.Sprintf("Ignore table '%s' (size %s B) %s", v, humanizeNumber(tableSize), useIcon))
				if defaultIgnore {
					item.Value = "checked"
				}
			}

			form.AddSubmit("Download SQL Backup")
		}, func(vc FormValidation, request *Request) {

			tables, err := listTables(app.db)
			must(err)

			var exclude []string

			for k := range tables {
				if request.Param(k) == "on" {
					exclude = append(exclude, k)
				}
			}

			excludes := strings.Join(exclude, ",")

			vc.Redirect("/admin/api/sqlbackup?exclude=" + excludes)
		}).Name(unlocalized("SQL Backup")).Permission(sysadminPermission).Board(sysadminBoard)

	app.API("sqlbackup").Permission("sysadmin").Handler(func(r *Request) {

		excludes := strings.Split(r.Param("exclude"), ",")

		r.Response().Header().Set("Content-Type", "application/octet-stream")
		filename := fmt.Sprintf("mysqldump_%s_%s.sql", app.codeName, time.Now().Format("2006-01-02_15:04:05"))
		r.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		err := app.backupSQL(r.Response(), excludes)
		if err != nil {
			app.Log().Printf("sqlbackup ended with error: %s", err)
		}
	})

}

func listTablesArr(app *App) []string {
	tables, err := listTables(app.db)
	must(err)

	var tablesAr []string

	for k := range tables {
		tablesAr = append(tablesAr, k)
	}

	sort.Strings(tablesAr)
	return tablesAr
}
