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
)

type BuildMiddleware struct {
	Copy [][2]string
}

func (b BuildMiddleware) Init(app *prago.App) error {

	var version = app.Data()["version"].(string)
	var appName = app.Data()["appName"].(string)
	ssh := app.Config().GetString("ssh")

	versionCommand := app.CreateCommand("version", "Print version")
	app.AddCommand(versionCommand, func(app *prago.App) error {
		fmt.Println(appName, version)
		return nil
	})

	buildCommand := app.CreateCommand("build", "Build cmd")
	app.AddCommand(buildCommand, func(app *prago.App) error {
		return b.build(appName, version)
	})

	releaseCommand := app.CreateCommand("release", "Release cmd")
	releaseCommandVersion := releaseCommand.Arg("version", "").Required().String()
	app.AddCommand(releaseCommand, func(app *prago.App) error {
		return b.release(appName, *releaseCommandVersion, ssh)
	})

	return nil
}

func (b BuildMiddleware) release(appName, version, auth string) error {
	from := os.Getenv("HOME") + "/." + appName + "/versions/" + appName + "." + version
	to := fmt.Sprintf("%s:~/.%s/versions", auth, appName)
	fmt.Println(from)
	fmt.Println(to)

	cmd := exec.Command("scp", "-r", from, to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type buildFlag struct {
	name   string
	goos   string
	goarch string
}

var linuxBuild = buildFlag{"linux", "linux", "386"}
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
	copyFiles(dirPath, buildPath)
	return nil
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
