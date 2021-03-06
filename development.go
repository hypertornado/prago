package prago

import (
	"os"
	"os/exec"
)

type DevelopmentSettings struct {
	Less       []Less
	TypeScript []string
}

type Less struct {
	SourceDir string
	Target    string
}

func (app *App) InitDevelopment(settings DevelopmentSettings) {
	var port int = 8585
	app.AddCommand("dev").
		Description("Development command").
		Flag(
			NewCommandFlag("port", "server port").
				Alias("p").
				Int(&port),
		).
		Callback(
			func() {
				app.DevelopmentMode = true
				for _, v := range settings.Less {
					go developmentLess(v.SourceDir, v.Target)
				}

				for _, v := range settings.TypeScript {
					go developmentTypescript(v)
				}

				err := app.ListenAndServe(port)
				if err != nil {
					panic(err)
				}
			})
}

func developmentTypescript(path string) {
	cmd := exec.Command("tsc", "-p", path, "-w")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
