package extensions

import (
	"fmt"
	"github.com/hypertornado/prago"
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

	versionCommand := app.CreateCommand("version", "Print version")
	app.AddCommand(versionCommand, func(app *prago.App) error {
		fmt.Println(appName, version)
		return nil
	})

	buildCommand := app.CreateCommand("build", "Build cmd")
	app.AddCommand(buildCommand, func(app *prago.App) error {
		return b.build(app.Data()["appName"].(string), "v1")
	})
	return nil
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

	//defer os.RemoveAll(dir)

	if true {
		for _, buildFlag := range []buildFlag{ /*linuxBuild,*/ macBuild} {
			err := buildExecutable(buildFlag, appName, dirPath)
			if err != nil {
				return err
			}
		}
	}

	for _, v := range b.Copy {
		copyPath := filepath.Join(dirPath, v[1])
		copyFiles(v[0], copyPath)
	}

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
