package prago

import (
	"embed"
	"errors"
	"fmt"
	"strings"
)

const iconResource = "glyphicons-basic-577-cluster.svg"
const iconTable = "glyphicons-basic-120-table.svg"
const iconAdd = "glyphicons-basic-371-plus.svg"
const iconSignpost = "glyphicons-basic-697-directions-sign.svg"
const iconBoard = "glyphicons-basic-424-blackboard.svg"
const iconForm = "glyphicons-basic-30-clipboard.svg"

func (app *App) SetIcons(iconsFS embed.FS, prefix string) {
	app.iconsFS = &iconsFS
	app.iconsPrefix = prefix
}

func (app *App) iconExists(iconName string) bool {
	if app.iconsFS == nil {
		return false
	}
	file, err := app.iconsFS.Open(app.iconsPrefix + iconName)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

func (app *App) loadIcon(iconName, color string) ([]byte, error) {

	if app.iconsFS == nil {
		return nil, errors.New("no icons set")
	}

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

	app.API("icons").Permission(everybodyPermission).Method("GET").Handler(func(request *Request) {
		if app.iconsFS == nil {
			request.WriteJSON(404, "icon not found")
			return
		}

		data, err := app.loadIcon(request.Param("file"), request.Param("color"))
		must(err)

		request.Response().Header().Add("Content-Type", "image/svg+xml")
		request.Response().Header().Add("Cache-Control", "max-age=604800")

		request.Response().Write(data)
	})

	app.Action("help/icons").Name(unlocalized("Ikony")).Permission(loggedPermission).hiddenInMenu().Template("admin_help_icons").DataSource(func(r *Request) interface{} {
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
