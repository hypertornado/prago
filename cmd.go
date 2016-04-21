package prago

import (
	"errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"os/exec"
	"strconv"
)

var defaultPort = 8585

type MiddlewareCmd struct{}

func (m MiddlewareCmd) Init(app *App) error {
	app.kingpin = kingpin.New("", "")
	app.commands = map[*kingpin.CmdClause]func(app *App) error{}

	devCommand := app.kingpin.Command("dev", "Development")
	app.commands[devCommand] = func(app *App) error {
		return development(app)
	}

	serverCommand := app.kingpin.Command("server", "Run server")
	portFlag := serverCommand.Flag("port", "server port").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()
	app.commands[serverCommand] = func(app *App) error {
		var port = defaultPort
		if portFlag != nil && *portFlag > 0 {
			port = *portFlag
		} else {
			config, err := app.Config()
			if err != nil {
				return err
			}
			configPort, ok := config["port"]
			if ok {
				port, err = strconv.Atoi(configPort)
				if err != nil {
					return errors.New("Wrong format of 'port' entry in config file. Should be int.")
				}
			}
		}
		return app.start(port, *developmentMode)
	}
	return nil
}

type MiddlewareRun struct{ Fn func(*App) }

func (mr MiddlewareRun) Init(app *App) error {
	mr.Fn(app)
	return nil
}

func (a *App) start(port int, developmentMode bool) error {
	return a.ListenAndServe(port, developmentMode)
}

func development(app *App) error {
	go developmentCSS()
	return app.start(defaultPort, true)
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
