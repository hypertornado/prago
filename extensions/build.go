package extensions

import (
	"fmt"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type BuildMiddleware struct{}

func (b BuildMiddleware) Init(app *prago.App) error {

	buildCommand := app.CreateCommand("build", "Build cmd")
	app.AddCommand(buildCommand, func(app *prago.App) error {
		return build(app.Data()["appName"].(string), "v1")
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

func build(appName, version string) error {
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

	fmt.Println(dirPath)

	if true {
		for _, buildFlag := range []buildFlag{ /*linuxBuild,*/ macBuild} {
			err := buildExecutable(buildFlag, appName, dirPath)
			if err != nil {
				return err
			}
		}
	}

	copyFiles("public", dirPath)

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
	return exec.Command("cp", "-R", from, to).Run()
}
