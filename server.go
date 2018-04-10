package prago

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"strconv"
)

var defaultPort = 8585

func initKingpinCommand(app *App) {
	app.kingpin = kingpin.New("", "")
	app.commands = map[*kingpin.CmdClause]func(app *App){}

	serverCommand := app.CreateCommand("server", "Run server")
	portFlag := serverCommand.Flag("port", "server port").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()

	app.AddCommand(serverCommand, func(app *App) {
		var port = defaultPort
		if portFlag != nil && *portFlag > 0 {
			port = *portFlag
		} else {
			configPort, err := app.Config.Get("port")
			if err == nil {
				port, err = strconv.Atoi(configPort.(string))
				if err != nil {
					app.Log().Fatalf("wrong format of 'port' entry in config file, should be int")
				}
			}
		}
		err := app.ListenAndServe(port, *developmentMode)
		if err != nil {
			app.Log().Fatal(err)
		}
	})
}
