package development

import (
	"github.com/hypertornado/prago"
)

func (m MiddlewareDevelopment) Init(app *prago.App) error {
	app.Log().Println("Development not implemented for linux")
	return nil
}
