package prago

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type development struct {
	app          *App
	less         []less
	sass         []sass
	typeScript   []string
	templateData []*developmentTemplateData
}

type less struct {
	SourceDir string
	Target    string
}

type sass struct {
	SourceDir string
	Target    string
}

type developmentTemplateData struct {
	Templates     *PragoTemplates
	WatchPath     string
	MatchPatterns []string
}

func (app *App) initDevelopment() {
	var dev = &development{
		app: app,
	}
	app.development = dev

	var port = defaultPort
	app.addCommand("dev").
		Description("Development command").
		flag(
			newCommandFlag("port", "server port").
				Alias("p").
				Int(&port),
		).
		Callback(
			func() {
				app.startDevelopment()
				err := app.listenAndServe(port)
				if err != nil {
					panic(err)
				}
			})
}

// AddTypeScriptDevelopmentPath automatically runs compilation of .tsc file in development mode
func (app *App) AddTypeScriptDevelopmentPath(path string) {
	app.development.typeScript = append(app.development.typeScript, path)
}

// AddLessDevelopmentPaths compiles less files in sourcePath into targetPath in development mode
func (app *App) AddLessDevelopmentPaths(sourcePath, targetPath string) {
	app.development.less = append(app.development.less, less{sourcePath, targetPath})
}

func (app *App) AddSassDevelopmentPaths(sourcePath, targetPath string) {
	app.development.sass = append(app.development.sass, sass{sourcePath, targetPath})
}

func (app *App) AddPragoDevelopmentPath(path string) {
	app.AddTemplatesDevelopmentPath(app.adminTemplates, path+"/templates", "*")
}

func (app *App) AddTemplatesDevelopmentPath(templates *PragoTemplates, watchPath string, matchPatterns ...string) {
	app.development.templateData = append(app.development.templateData, &developmentTemplateData{
		WatchPath:     watchPath,
		MatchPatterns: matchPatterns,
		Templates:     templates,
	})
}

func (app *App) startDevelopment() {
	app.developmentMode = true
	for _, v := range app.development.less {
		go app.developmentLess(v.SourceDir, v.Target)
	}

	for _, v := range app.development.sass {
		go app.developmentSass(v.SourceDir, v.Target)
	}

	for _, v := range app.development.typeScript {
		go developmentTypescript(v)
	}

	for _, v := range app.development.templateData {
		app.developmentTemplate(v)
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
	watchPath(sourcePath, func() {
		compileLess(indexPath, targetPath)
	})
}

func (app *App) developmentSass(sourcePath, targetPath string) {
	indexPath := filepath.Join(sourcePath, "index.scss")
	compileSass(indexPath, targetPath)
	watchPath(sourcePath, func() {
		compileSass(indexPath, targetPath)
	})
}

func (app *App) developmentTemplate(data *developmentTemplateData) {

	data.Templates.watchPattern = data.WatchPath
	data.Templates.matchPatterns = data.MatchPatterns
	data.Templates.fs = os.DirFS(data.WatchPath)
	go data.Templates.watch()

	//must(app.adminTemplates.SetFilesystem(os.DirFS(data.WatchPath), data.MatchPatterns...))

	/*watchPath(data.WatchPath, func() {
		data.Templates.templatesMutex.Lock()
		defer data.Templates.templatesMutex.Unlock()
		app.Log().Printf("Compiling changed templates from path: %s", data.WatchPath)
		err := app.adminTemplates.parseTemplates()
		if err != nil {
			app.Log().Printf("Error while compiling templates in development mode from path '%s': %s", data.WatchPath, err)
		} else {
			app.Log().Println("Compiling OK.")
		}
	})*/
}

func compileLess(from, to string) error {
	outfile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("lessc", from), outfile)
}

func compileSass(from, to string) error {
	outfile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("sass", from), outfile)
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
