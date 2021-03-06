package prago

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
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
					go app.developmentLess(v.SourceDir, v.Target)
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

func (app *App) fileWatcher(path string, handler func()) {
	w := watcher.New()
	w.SetMaxEvents(1)

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
				handler()
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(path); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		panic(err)
	}
}

func (app *App) developmentLess(sourcePath, targetPath string) {
	compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
	app.fileWatcher(sourcePath, func() {
		compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
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
