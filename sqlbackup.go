package prago

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (app *App) initSQLBackup() {

	app.FormAction("sqlbackup",
		func(form *Form, request *Request) {
			form.Title = "SQL Backup"

			tables, err := listTables(app.db)
			must(err)

			var tablesAr []string

			for k := range tables {
				tablesAr = append(tablesAr, k)
			}

			sort.Strings(tablesAr)

			for _, v := range tablesAr {
				form.AddCheckbox(v, fmt.Sprintf("Ignore table '%s'", v))
			}

			form.AddSubmit("Download SQL Backup")
		}, func(vc ValidationContext) {

			tables, err := listTables(app.db)
			must(err)

			var exclude []string

			for k := range tables {
				if vc.GetValue(k) == "on" {
					exclude = append(exclude, k)
				}
			}

			excludes := strings.Join(exclude, ",")

			vc.Validation().RedirectionLocaliton = "/admin/api/sqlbackup?exclude=" + excludes
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
