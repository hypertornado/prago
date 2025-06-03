package prago

type AppInfo struct {
	CodeName string
	Name     func(string) string
	Version  string
	HasLogo  bool
}

func (app *App) GetAppInfo() AppInfo {
	var hasLogo bool
	if app.logo != nil {
		hasLogo = true
	}

	return AppInfo{
		CodeName: app.codeName,
		Name:     app.name,
		Version:  app.version,
		HasLogo:  hasLogo,
	}
}
