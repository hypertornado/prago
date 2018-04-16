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
	var port int = 8585
	app.AddCommand("dev").
		Description("Development command").
		Flag(
			prago.NewFlag("port", "server port").
				Alias("p").
				Int(&port),
		).
		Callback(
			func() {
				for _, v := range settings.Less {
					go developmentLess(v.SourceDir, v.Target)
				}

				for _, v := range settings.TypeScript {
					go developmentTypescript(v)
				}

				err := app.ListenAndServe(port, true)
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