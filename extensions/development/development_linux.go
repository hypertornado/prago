package development

import (
	"github.com/hypertornado/prago"
)

func (m MiddlewareDevelopment) initPlatform(app *prago.App) error {
	app.Log().Println("Development not implemented for linux")
	return nil
}
