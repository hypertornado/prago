package prago

/*type Tab struct {
	Icon     string
	Name     string
	URL      string
	Selected bool
}*/

type Button struct {
	Icon     string
	Name     string
	URL      string
	Selected bool
}

func (app *App) getLoginNavigation(language, code string) (ret []*Button) {
	ret = append(ret, &Button{
		Name:     messages.Get(language, "admin_login_action"),
		Icon:     "glyphicons-basic-431-log-in.svg",
		URL:      app.getAdminURL("user/login"),
		Selected: code == "login",
	})

	ret = append(ret, &Button{
		Name:     messages.Get(language, "admin_register"),
		Icon:     "glyphicons-basic-7-user-plus.svg",
		URL:      app.getAdminURL("user/registration"),
		Selected: code == "registration",
	})

	ret = append(ret, &Button{
		Name:     messages.Get(language, "admin_forgotten"),
		Icon:     "glyphicons-basic-45-key.svg",
		URL:      app.getAdminURL("user/forgot"),
		Selected: code == "forgot",
	})
	return
}
