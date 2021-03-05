package prago

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func (app *App) initBackupCRON() {
	app.NewTask("backup_db").SetHandler(
		func(tr *TaskActivity) error {
			err := BackupApp(app)
			if err != nil {
				return fmt.Errorf("Error while creating backup: %s", err)
			}
			return nil
		}).RepeatEvery(24 * time.Hour)

	app.NewTask("remove_old_backups").SetHandler(
		func(tr *TaskActivity) error {
			tr.SetStatus(0, fmt.Sprintf("Removing old backups"))
			deadline := time.Now().AddDate(0, 0, -7)
			backupPath := app.dotPath() + "/backups"
			files, err := ioutil.ReadDir(backupPath)
			if err != nil {
				return fmt.Errorf("Error while removing old backups: %s", err)
			}
			for _, file := range files {
				if file.ModTime().Before(deadline) {
					removePath := backupPath + "/" + file.Name()
					err := os.RemoveAll(removePath)
					if err != nil {
						return fmt.Errorf("Error while removing old backup file: %s", err)
					}
				}
			}
			app.Log().Println("Old backups removed")
			return nil
		}).RepeatEvery(1 * time.Hour)

}
