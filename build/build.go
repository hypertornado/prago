package build

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
type BuildSettings struct {
	Copy [][2]string
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

//Init initializes build middleware
func CreateBuildHelper(app *prago.App, b BuildSettings) {
	var version = app.Version
	var appName = app.AppName

	app.AddCommand("build").Callback(func() {
		b.build(appName, version)
	})

	ssh := app.Config.GetStringWithFallback("ssh", "")
	if ssh == "" {
		app.Log().Error("no ssh value set in config file")
		return
	}

	app.AddCommand("backup").Callback(func() {
		must(BackupApp(app))
	})

	app.AddCommand("syncbackups").Callback(func() {
		must(b.syncBackups(appName, ssh))
	})

	app.AddCommand("party").Callback(func() {
		must(b.party(appName, version, ssh))
	})

}

func (b BuildSettings) party(appName, version, ssh string) (err error) {
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

func (b BuildSettings) syncBackups(appName, ssh string) error {
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

func (b BuildSettings) remote(appName, version, ssh string) error {
	cmdStr := fmt.Sprintf("cd ~/.%s/versions/%s.%s; ./%s.linux admin migrate; killall %s.linux; nohup ./%s.linux server >> app.log 2>&1 & exit;", appName, appName, version, appName, appName, appName)
	println(cmdStr)
	cmd := exec.Command("ssh", ssh, cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

//BackupApp backups whole app
func BackupApp(app *prago.App) error {

	app.Log().Println("Creating backup")

	var appName = app.AppName

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

	user := app.Config.GetString("dbUser")
	dbName := app.Config.GetString("dbName")
	password := app.Config.GetString("dbPassword")

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

	//TODO: enable backup of static resources
	//paths, err := app.Config.Get("staticPaths")
	/*if err == nil {
		for k, v := range paths.([]interface{}) {

			staticPath := filepath.Join(dirPath, "static", fmt.Sprintf("%d", k))

			err = exec.Command("mkdir", "-p", staticPath).Run()
			if err != nil {
				return fmt.Errorf("mkdir for static paths backup: %s", err)
			}

			err = copyFiles(v.(string), staticPath)
			if err != nil {
				return fmt.Errorf("copying backup files: %s", err)
			}
		}
	}*/

	backupsPath := filepath.Join(os.Getenv("HOME"), "."+appName, "backups")
	err = exec.Command("mkdir", "-p", backupsPath).Run()
	if err != nil {
		return fmt.Errorf("making dir for backup files: %s", err)
	}

	return copyFiles(dirPath, backupsPath)
}

func (b BuildSettings) release(appName, version, ssh string) error {
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

func (b BuildSettings) build(appName, version string) error {
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

	for _, buildFlag := range []buildFlag{linuxBuild} {
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
	err := exec.Command("cp", "-R", from, to).Run()
	if err != nil {
		return fmt.Errorf("error while copying files from %s to %s: %s", from, to, err)
	}
	return nil
}
