package prago

type UserData interface {
	Name() string
	Locale() string
	Authorize(Permission) bool
}

type userData struct {
	name   string
	role   string
	locale string
	app    *App
}

func (app *App) newUserData(user *user) *userData {
	return &userData{
		name:   user.LongName(),
		role:   user.Role,
		locale: user.Locale,
		app:    app,
	}
}

func (d *userData) Name() string {
	return d.name
}

func (d *userData) Locale() string {
	return d.locale
}

func (d *userData) Authorize(permission Permission) bool {
	return d.app.authorize(true, d.role, permission)
}
