package prago

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (app *App) initBackupCRON() {

	backupDashboard := sysadminBoard.Dashboard(unlocalized("Backup"))

	app.addCronTask("backup db", 24*time.Hour, func() {
		err := backupApp(app)
		if err != nil {
			app.Log().Errorf("error while creating backup: %s", err)
		}
	})

	backupDashboard.
		AddTask(unlocalized("backup_db"), "sysadmin",
			func(tr *TaskActivity) error {
				err := backupApp(app)
				if err != nil {
					return fmt.Errorf("error while creating backup: %s", err)
				}
				return nil
			})

	app.addCronTask("remove old backups", 24*time.Hour, func() {
		err := app.removeOldBackups()
		if err != nil {
			app.Log().Errorf("error while removing old backups: %s", err)
		}
	})

	backupDashboard.AddTask(unlocalized("remove_old_backups"), "sysadmin",
		func(ta *TaskActivity) error {
			return app.removeOldBackups()
		})
}

func (app *App) removeOldBackups() error {
	app.Log().Println("Removing old backups")
	deadline := time.Now().AddDate(0, 0, -7)
	backupPath := app.dotPath() + "/backups"
	files, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("error while removing old backups: %s", err)
	}
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return fmt.Errorf("can't get file info: %s", err)
		}
		if info.ModTime().Before(deadline) {
			removePath := backupPath + "/" + file.Name()
			err := os.RemoveAll(removePath)
			if err != nil {
				return fmt.Errorf("error while removing old backup file: %s", err)
			}
		}
	}
	app.Log().Println("Old backups removed")
	return nil

}

func backupApp(app *App) error {
	app.Log().Println("Creating backup")
	var appName = app.codeName

	dir, err := os.MkdirTemp("", "backup")
	if err != nil {
		return fmt.Errorf("creating backup tmp dir: %s", err)
	}

	dirPath := filepath.Join(dir, time.Now().Format("2006-01-02_15:04:05"))
	err = os.Mkdir(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("creating backup tmp dir with date: %s", err)
	}
	defer os.RemoveAll(dir)

	dbFilePath := filepath.Join(dirPath, "db.sql")

	dbFile, err := os.Create(dbFilePath)
	if err != nil {
		return fmt.Errorf("creating backup db file: %s", err)
	}
	defer dbFile.Close()

	err = app.backupSQL(dbFile, nil)
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

func (app *App) restoreSQLBackup(reader io.Reader) error {

	password := app.dbConfig.Password
	var command *exec.Cmd
	var params []string
	if password == "" {
		params = []string{"-u" + app.dbConfig.User, app.dbConfig.Name}
	} else {
		params = []string{"-u" + app.dbConfig.User, "-p" + password, app.dbConfig.Name}
	}
	app.Log().Printf("mysql with params %v", params)
	command = exec.Command("mysql", params...)
	command.Stdout = os.Stdout
	command.Stdin = reader
	return command.Run()

}

func (app *App) backupSQL(writer io.Writer, excludeTableNames []string) error {

	availableTables, err := listTables(app.db)
	if err != nil {
		return err
	}

	var appendedIgnore []string

	for _, v := range excludeTableNames {
		if v == "" {
			continue
		}
		if !availableTables[v] {
			return fmt.Errorf("can't ignore table '%s'", v)
		}
		appendedIgnore = append(appendedIgnore, fmt.Sprintf("--ignore-table=%s.%s", app.dbConfig.Name, v))

	}

	password := app.dbConfig.Password
	var dumpCmd *exec.Cmd
	var params []string
	if password == "" {
		params = []string{"-u" + app.dbConfig.User, app.dbConfig.Name}
	} else {
		params = []string{"-u" + app.dbConfig.User, "-p" + password, app.dbConfig.Name}

	}
	params = append(params, appendedIgnore...)
	app.Log().Printf("mysqldump with params %v", params)
	dumpCmd = exec.Command("mysqldump", params...)
	dumpCmd.Stdout = writer
	return dumpCmd.Run()
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
