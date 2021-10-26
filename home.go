package prago

func (app *App) initHome() {
	app.
		Action("").
		Permission(loggedPermission).
		Name(messages.GetNameFunction("admin_signpost")).
		Template("admin_home_navigation").
		DataSource(app.getHomeData).
		IsWide()
}

func (app *App) getHomeData(request *Request) interface{} {
	return app.getMainMenu(request)

}
