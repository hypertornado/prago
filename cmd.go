package prago

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"os/exec"
)

type MiddlewareCmd struct{}

func (MiddlewareCmd) Init(app *App) error {
	println("init")
	return nil
}

func (app *App) cmd(init func(*App)) error {
	cmd := kingpin.New("", "")
	serverCommand := cmd.Command("server", "Run server")
	port := serverCommand.Flag("port", "server port").Default("8585").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()

	buildCommand := cmd.Command("build", "Build version")

	cssCommand := cmd.Command("css", "Build CSS")
	devCommand := cmd.Command("dev", "Development")

	command, err := cmd.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	switch command {
	case serverCommand.FullCommand():
		init(app)
		app.start(*port, *developmentMode)
	case buildCommand.FullCommand():
		build()
	case cssCommand.FullCommand():
		compileCss()
	case devCommand.FullCommand():
		init(app)
		development(app)
	}
	return nil
}

func (a *App) start(port int, developmentMode bool) {
	err := a.ListenAndServe(port, developmentMode)
	if err != nil {
		panic(err)
	}
}

func build() {
	println("not implemented")
}

func development(app *App) {
	go developmentCSS()
	app.start(8585, true)
}

func compileCss() error {
	outfile, err := os.Create("public/compiled.css")
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("lessc", "public/css/index.less"), outfile)
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

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
