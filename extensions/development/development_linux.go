package development

type DevelopmentSettings struct{}

type MiddlewareDevelopment struct {
	Settings DevelopmentSettings
}

func (m MiddlewareDevelopment) Init(app *prago.App) error {
	panic("development not implemented for linux")
	return nil
}
