package prago

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type DevelopmentSettings struct {
	Less       []Less
	TypeScript []string
}

type Less struct {
	SourceDir string
	Target    string
}

type developmentTemplatePath struct {
	Path     string
	Patterns []string
}

type development struct {
	app           *App
	Less          []Less
	TypeScript    []string
	templatePaths []developmentTemplatePath
}

func (app *App) initDevelopment() {
	var dev = &development{
		app: app,
	}
	app.development = dev

	var port int = defaultPort
	app.AddCommand("dev").
		Description("Development command").
		Flag(
			NewCommandFlag("port", "server port").
				Alias("p").
				Int(&port),
		).
		Callback(
			func() {
				app.startDevelopment()
				err := app.ListenAndServe(port)
				if err != nil {
					panic(err)
				}
			})
}

func (app *App) AddTypeScriptDevelopmentPath(path string) {
	app.development.TypeScript = append(app.development.TypeScript, path)
}

func (app *App) AddLessDevelopmentPaths(sourcePath, targetPath string) {
	app.development.Less = append(app.development.Less, Less{sourcePath, targetPath})
}

func (app *App) AddTemplatesDevelopmentPath(path string, patterns ...string) {
	app.development.templatePaths = append(app.development.templatePaths, developmentTemplatePath{
		Path:     path,
		Patterns: patterns,
	})
}

func (app *App) startDevelopment() {
	app.DevelopmentMode = true
	for _, v := range app.development.Less {
		go app.developmentLess(v.SourceDir, v.Target)
	}

	for _, v := range app.development.TypeScript {
		go developmentTypescript(v)
	}

	for _, v := range app.development.templatePaths {
		go app.developmentTemplate(v)
	}
}

func developmentTypescript(path string) {
	cmd := exec.Command("tsc", "-p", path, "-w")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}

func (app *App) developmentLess(sourcePath, targetPath string) {
	indexPath := filepath.Join(sourcePath, "index.less")
	compileLess(indexPath, targetPath)
	app.watchPath(sourcePath, func() {
		compileLess(indexPath, targetPath)
	})
}

func (app *App) developmentTemplate(path developmentTemplatePath) {
	must(app.AddTemplates(os.DirFS(path.Path), path.Patterns...))

	app.watchPath(path.Path, func() {
		app.Log().Printf("Compiling changed templates from path: %s", path.Path)
		err := app.parseTemplates()
		if err != nil {
			app.Log().Printf("Error while compiling templates in development mode from path '%s': %s", path.Path, err)
		} else {
			app.Log().Println("Compiling OK.")
		}
	})
}

func compileLess(from, to string) error {
	outfile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("lessc", from), outfile)
}

func commandHelper(cmd *exec.Cmd, out io.Writer) error {
	var err error
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
