package development

import (
	"github.com/hypertornado/prago"
	"os"
	"os/exec"
)

var defaultPort = 8585

type DevelopmentSettings struct {
	Less       []Less
	TypeScript []string
}

type Less struct {
	SourceDir string
	Target    string
}

func CreateDevelopmentHelper(app *prago.App, settings DevelopmentSettings) {
	devCommand := app.CreateCommand("dev", "Development")
	portFlag := devCommand.Flag("port", "server port").Short('p').Default("8585").Int()
	developmentMode := devCommand.Flag("development", "Is in development mode").Default("t").Short('d').Bool()

	app.AddCommand(devCommand, func(app *prago.App) error {
		for _, v := range settings.Less {
			go developmentLess(v.SourceDir, v.Target)
		}

		for _, v := range settings.TypeScript {
			go developmentTypescript(v)
		}

		return app.ListenAndServe(*portFlag, *developmentMode)
	})
}

func developmentTypescript(path string) {
	cmd := exec.Command("tsc", "-p", path, "-w")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
