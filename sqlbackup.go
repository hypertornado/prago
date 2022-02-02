package prago

import (
	"fmt"
	"time"
)

func (app *App) initSQLBackup() {

	app.FormAction("sqlbackup").Name(unlocalized("SQL Backup")).Permission(sysadminPermission).Form(
		func(form *Form, request *Request) {
			form.Title = "SQL Backup"
			form.AddSubmit("Download SQL Backup")
		},
	).Validation(func(vc ValidationContext) {
		vc.Validation().RedirectionLocaliton = "/admin/api/sqlbackup"
	})

	app.API("sqlbackup").Permission("sysadmin").Handler(func(r *Request) {
		r.Response().Header().Set("Content-Type", "application/octet-stream")
		filename := fmt.Sprintf("mysqldump_%s_%s.sql", app.codeName, time.Now().Format("2006-01-02_15:04:05"))
		r.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		err := app.backupSQL(r.Response(), r.Request().Context())
		if err != nil {
			app.Log().Printf("sqlbackup ended with error: %s", err)
		}
	})

}
