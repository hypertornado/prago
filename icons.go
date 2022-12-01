package prago

import (
	"embed"
	"fmt"
	"strings"
)

const iconResource = "glyphicons-basic-577-cluster.svg"
const iconTable = "glyphicons-basic-120-table.svg"
const iconAdd = "glyphicons-basic-371-plus.svg"

func (app *App) SetIcons(iconsFS embed.FS, prefix string) {
	app.iconsFS = &iconsFS
	app.iconsPrefix = prefix

}

func (app *App) loadIcon(iconName, color string) ([]byte, error) {
	data, err := app.iconsFS.ReadFile(app.iconsPrefix + iconName)
	if err != nil {
		return nil, err
	}

	str := string(data)

	if color != "" {
		str = fmt.Sprintf("<svg fill=\"#%s\" %s", color, str[4:])
	}

	return []byte(str), nil

}

func (app *App) initIcons() {
	app.API("icons").Permission(loggedPermission).Method("GET").Handler(func(request *Request) {
		if app.iconsFS == nil {
			panic("no icons fs set")
		}

		data, err := app.loadIcon(request.Param("file"), request.Param("color"))
		must(err)

		request.Response().Header().Add("Content-Type", "image/svg+xml")
		request.Response().Write(data)

	})

	app.Action("help/icons").Name(unlocalized("Ikony")).Permission(loggedPermission).hiddenInMainMenu().Template("admin_help_icons").DataSource(func(r *Request) interface{} {
		prefix := app.iconsPrefix
		prefix = strings.TrimRight(prefix, "/")

		icons, err := app.iconsFS.ReadDir(prefix)
		must(err)

		var ret []string

		for _, v := range icons {
			ret = append(ret, v.Name())
		}
		return ret
	})

}
