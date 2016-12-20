package extensions

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

//BuildMiddleware allows binary building and release
type BuildMiddleware struct {
	Copy [][2]string
}

//Init initializes build middleware
func (b BuildMiddleware) Init(app *prago.App) error {

	var version = app.Data()["version"].(string)
	var appName = app.Data()["appName"].(string)

	versionCommand := app.CreateCommand("version", "Print version")
	app.AddCommand(versionCommand, func(app *prago.App) error {
		fmt.Println(appName, version)
		return nil
	})

	buildCommand := app.CreateCommand("build", "Build cmd")
	app.AddCommand(buildCommand, func(app *prago.App) error {
		return b.build(appName, version)
	})

	sshVal, err := app.Config.Get("ssh")

	if err == nil {
		ssh := sshVal.(string)

		releaseCommand := app.CreateCommand("release", "Release cmd")
		releaseCommandVersion := releaseCommand.Arg("version", "").Default(version).String()
		app.AddCommand(releaseCommand, func(app *prago.App) error {
			return b.release(appName, *releaseCommandVersion, ssh)
		})

		remoteCommand := app.CreateCommand("remote", "Remote")
		remoteCommandVersion := remoteCommand.Arg("version", "").Default(version).String()
		app.AddCommand(remoteCommand, func(app *prago.App) error {
			return b.remote(appName, *remoteCommandVersion, ssh)
		})

		backupCommand := app.CreateCommand("backup", "Backup")
		app.AddCommand(backupCommand, BackupApp)

		syncBackupCommand := app.CreateCommand("syncbackups", "Sync backups from server")
		app.AddCommand(syncBackupCommand, func(app *prago.App) error {
			return b.syncBackups(appName, ssh)
		})

		partyCommand := app.CreateCommand("party", "release and run current version")
		app.AddCommand(partyCommand, func(app *prago.App) error {
			return b.party(appName, version, ssh)
		})

	}

	return nil
}

func (b BuildMiddleware) party(appName, version, ssh string) (err error) {
	if err = b.build(appName, version); err != nil {
		return err
	}
	if err = b.release(appName, version, ssh); err != nil {
		return err
	}
	if err = b.remote(appName, version, ssh); err != nil {
		return err
	}
	return nil
}

func (b BuildMiddleware) syncBackups(appName, ssh string) error {
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

func (b BuildMiddleware) remote(appName, version, ssh string) error {
	cmdStr := fmt.Sprintf("cd ~/.%s/versions/%s.%s; ./%s.linux admin migrate; killall %s.linux; nohup ./%s.linux server & exit;", appName, appName, version, appName, appName, appName)
	println(cmdStr)
	cmd := exec.Command("ssh", ssh, cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

//BackupApp backups whole app
func BackupApp(app *prago.App) error {
	app.Log().Println("Creating backup")

	var appName = app.Data()["appName"].(string)

	dir, err := ioutil.TempDir("", "backup")
	if err != nil {
		return err
	}

	dirPath := filepath.Join(dir, time.Now().Format("2006-01-02_15:04:05"))
	err = os.Mkdir(dirPath, 0777)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	user := app.Config.GetString("dbUser")
	dbName := app.Config.GetString("dbName")
	password := app.Config.GetString("dbPassword")

	cmd := exec.Command("mysqldump", "-u"+user, "-p"+password, dbName)

	dbFilePath := filepath.Join(dirPath, "db.sql")

	dbFile, err := os.Create(dbFilePath)
	defer dbFile.Close()
	if err != nil {
		return err
	}

	cmd.Stdout = dbFile

	err = cmd.Run()
	if err != nil {
		return err
	}

	paths, err := app.Config.Get("staticPaths")

	if err == nil {
		for k, v := range paths.([]interface{}) {

			staticPath := filepath.Join(dirPath, "static", fmt.Sprintf("%d", k))

			err = exec.Command("mkdir", "-p", staticPath).Run()
			if err != nil {
				return err
			}

			err = copyFiles(v.(string), staticPath)
			if err != nil {
				return err
			}
		}
	}

	backupsPath := filepath.Join(os.Getenv("HOME"), "."+appName, "backups")
	err = exec.Command("mkdir", "-p", backupsPath).Run()
	if err != nil {
		return err
	}

	return copyFiles(dirPath, backupsPath)
}

func (b BuildMiddleware) release(appName, version, ssh string) error {
	from := os.Getenv("HOME") + "/." + appName + "/versions/" + appName + "." + version
	to := fmt.Sprintf("%s:~/.%s/versions", ssh, appName)

	mkdirStr := fmt.Sprintf("mkdir -p ~/.%s/versions", appName)
	fmt.Println(mkdirStr)
	cmd := exec.Command("ssh", ssh, mkdirStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("scp", "-r", from, to)
	cmd = exec.Command("scp", "-r", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type buildFlag struct {
	name   string
	goos   string
	goarch string
}

var linuxBuild = buildFlag{"linux", "linux", "amd64"}
var macBuild = buildFlag{"mac", "darwin", "amd64"}

func (b BuildMiddleware) build(appName, version string) error {
	fmt.Println(appName, version)
	dir, err := ioutil.TempDir("", "build")
	if err != nil {
		return err
	}

	dirName := fmt.Sprintf("%s.%s", appName, version)
	dirPath := filepath.Join(dir, dirName)
	err = os.Mkdir(dirPath, 0777)
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)

	for _, buildFlag := range []buildFlag{linuxBuild, macBuild} {
		err := buildExecutable(buildFlag, appName, dirPath)
		if err != nil {
			return err
		}
	}

	for _, v := range b.Copy {
		copyPath := filepath.Join(dirPath, v[1])
		copyFiles(v[0], copyPath)
	}

	buildPath := os.Getenv("HOME") + "/." + appName + "/versions"
	os.Mkdir(buildPath, 0777)
	buildDir := buildPath + "/" + dirName

	_, err = os.Open(buildDir)
	if err == nil {
		question := fmt.Sprintf("There is already file '%s'. Do you want to delete?", buildDir)
		if utils.ConsoleQuestion(question) {
			fmt.Println("Deleting " + buildDir)
			os.RemoveAll(buildDir)
		} else {
			return errors.New("Have not deleted old version.")
		}
	}
	return copyFiles(dirPath, buildPath)
}

func buildExecutable(bf buildFlag, appName, dirPath string) error {
	executablePath := filepath.Join(dirPath, fmt.Sprintf("%s.%s", appName, bf.name))
	fmt.Println("building", bf.name, "at", executablePath)
	cmd := exec.Command("go", "build", "-o", executablePath)
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", bf.goos))
	env = append(env, fmt.Sprintf("GOARCH=%s", bf.goarch))
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFiles(from, to string) error {
	fmt.Println("copying", from, "to", to)
	return exec.Command("cp", "-R", from, to).Run()
}
