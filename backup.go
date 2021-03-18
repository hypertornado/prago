package prago

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (app *App) initBackupCRON() {
	app.
		taskManager.
		defaultGroup.
		Task("backup_db").
		Handler(
			func(tr *TaskActivity) error {
				err := backupApp(app)
				if err != nil {
					return fmt.Errorf("Error while creating backup: %s", err)
				}
				return nil
			}).RepeatEvery(24 * time.Hour)

	app.taskManager.defaultGroup.Task("remove_old_backups").Handler(
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

func backupApp(app *App) error {

	app.Log().Println("Creating backup")

	var appName = app.codeName

	dir, err := ioutil.TempDir("", "backup")
	if err != nil {
		return fmt.Errorf("creating backup tmp dir: %s", err)
	}

	dirPath := filepath.Join(dir, time.Now().Format("2006-01-02_15:04:05"))
	err = os.Mkdir(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("creating backup tmp dir with date: %s", err)
	}
	defer os.RemoveAll(dir)

	user := app.ConfigurationGetStringWithFallback("dbUser", "")
	dbName := app.ConfigurationGetStringWithFallback("dbName", "")
	password := app.ConfigurationGetStringWithFallback("dbPassword", "")

	var dumpCmd *exec.Cmd

	if password == "" {
		dumpCmd = exec.Command("mysqldump", "-u"+user, dbName)
	} else {
		dumpCmd = exec.Command("mysqldump", "-u"+user, "-p"+password, dbName)
	}

	dbFilePath := filepath.Join(dirPath, "db.sql")

	dbFile, err := os.Create(dbFilePath)
	defer dbFile.Close()
	if err != nil {
		return fmt.Errorf("creating backup db file: %s", err)
	}

	dumpCmd.Stdout = dbFile

	err = dumpCmd.Run()
	if err != nil {
		return fmt.Errorf("dumping cmd: %s", err)
	}

	backupsPath := filepath.Join(os.Getenv("HOME"), "."+appName, "backups")
	err = exec.Command("mkdir", "-p", backupsPath).Run()
	if err != nil {
		return fmt.Errorf("making dir for backup files: %s", err)
	}

	return copyFiles(dirPath, backupsPath)
}

func syncBackups(appName, ssh string) error {
	to := filepath.Join(os.Getenv("HOME"), "."+appName, "serverbackups")
	err := exec.Command("mkdir", "-p", to).Run()
	if err != nil {
		return err
	}

	from := fmt.Sprintf("%s:~/.%s/backups/*", ssh, appName)

	fmt.Println("scp", "-r", from, to)
	cmd := exec.Command("scp", "-r", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
