package prago

import (
	"errors"
	"gopkg.in/alecthomas/kingpin.v2"
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
			configPort, err := app.Config().Get("port")
			if err == nil {
				port, err = strconv.Atoi(configPort.(string))
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
	_, ok := app.data["development"]
	if ok {
		fn, ok := app.data["development"].(func())
		if ok {
			go fn()
		}
	}
	return app.start(defaultPort, true)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
