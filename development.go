package prago

func development(app *App) error {
	_, ok := app.data["development"]
	if ok {
		fn, ok := app.data["development"].(func())
		if ok {
			go fn()
		}
	}
	return app.StartServer(defaultPort, true)
}

type MiddlewareDevelopment struct{}

func (m MiddlewareDevelopment) Init(app *App) error {
	app.AddCommand(app.kingpin.Command("dev", "Development"), development)
	return nil
}
