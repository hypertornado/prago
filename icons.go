package prago

import (
	"embed"
	"fmt"
)

func (app *App) SetIcons(iconsFS embed.FS, prefix string) {
	app.iconsFS = &iconsFS
	app.iconsPrefix = prefix

}

func (app *App) initIcons() {

	app.API("icons").Permission(loggedPermission).Method("GET").Handler(func(request *Request) {
		if app.iconsFS == nil {
			panic("no icons fs set")
		}

		data, err := app.iconsFS.ReadFile(app.iconsPrefix + request.Param("file"))
		if err != nil {
			must(err)
		}

		str := string(data)

		color := request.Param("color")
		if color != "" {
			str = fmt.Sprintf("<svg fill=\"#%s\" %s", color, str[4:])
		}

		request.Response().Header().Add("Content-Type", "image/svg+xml")
		request.Response().Write([]byte(str))

	})

}
