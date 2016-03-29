package prago

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"os/exec"
)

type MiddlewareCmd struct {
	kingpin *kingpin.Application
}

func (m MiddlewareCmd) Init(app *App) error {
	cmd := kingpin.New("", "")

	xx := cmd.Command("xx", "xx help")

	commands := map[*kingpin.CmdClause]func(app *App) error{}

	commands[xx] = func(app *App) error {
		println("xx cmd")
		return nil
	}

	app.data["kingpin"] = cmd
	app.data["commands"] = commands
	return nil
}

type MiddlewareRun struct {
	Fn func(*App)
}

func (mr MiddlewareRun) Init(app *App) error {
	app.kingpin = kingpin.New("", "")
	app.commands = map[*kingpin.CmdClause]func(app *App) error{}

	devCommand := app.kingpin.Command("dev", "Development")
	app.commands[devCommand] = func(app *App) error {
		mr.Fn(app)
		return development(app)
	}

	serverCommand := app.kingpin.Command("server", "Run server")
	port := serverCommand.Flag("port", "server port").Default("8585").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()
	app.commands[serverCommand] = func(app *App) error {
		mr.Fn(app)
		return app.start(*port, *developmentMode)
	}

	return nil
}

func (mr MiddlewareRun) InitOLD(app *App) error {
	cmd := app.data["kingpin"].(*kingpin.Application)
	serverCommand := cmd.Command("server", "Run server")
	port := serverCommand.Flag("port", "server port").Default("8585").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()

	devCommand := cmd.Command("dev", "Development")

	commandName, err := cmd.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	switch commandName {
	case serverCommand.FullCommand():
		mr.Fn(app)
		app.start(*port, *developmentMode)
	case devCommand.FullCommand():
		mr.Fn(app)
		development(app)
	default:
		commands := app.data["commands"].(map[*kingpin.CmdClause]func(app *App) error)
		for command, fn := range commands {
			if command.FullCommand() == commandName {
				return fn(app)
			}
		}
	}
	return nil
}

func (a *App) start(port int, developmentMode bool) error {
	return a.ListenAndServe(port, developmentMode)
}

func development(app *App) error {
	go developmentCSS()
	app.start(8585, true)
	return nil
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
