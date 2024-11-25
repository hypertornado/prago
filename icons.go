package prago

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"
)

// const iconResource = "glyphicons-basic-577-cluster.svg"
const iconResource = "glyphicons-basic-964-layers.svg"
const iconTable = "glyphicons-basic-120-table.svg"
const iconAdd = "glyphicons-basic-371-plus.svg"
const iconSignpost = "glyphicons-basic-697-directions-sign.svg"
const iconBoard = "glyphicons-basic-424-blackboard.svg"
const iconForm = "glyphicons-basic-30-clipboard.svg"
const iconView = "glyphicons-basic-588-book-open-text.svg"
const iconAction = "glyphicons-basic-111-paragraph-left.svg"
const iconEdit = "glyphicons-basic-31-pencil.svg"
const iconDelete = "glyphicons-basic-17-bin.svg"
const iconDuplicate = "glyphicons-basic-611-copy-duplicate.svg"
const iconActivity = "glyphicons-basic-58-history.svg"

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

func iconColorPrefix(color string) io.Reader {
	if len(color) != 6 {
		return strings.NewReader("<svg ")
	}
	return strings.NewReader(fmt.Sprintf("<svg fill=\"#%s\" ", color))
}

func (app *App) initIcons() {

	app.API("icons").Permission(everybodyPermission).Method("GET").Handler(func(request *Request) {
		if app.iconsFS == nil {
			request.WriteJSON(404, "icon not found")
			return
		}

		request.Response().Header().Add("Content-Type", "image/svg+xml")
		request.Response().Header().Add("Cache-Control", "max-age=604800")

		file, err := app.iconsFS.Open(app.iconsPrefix + request.Param("file"))
		must(err)
		defer file.Close()

		io.Copy(request.Response(), iconColorPrefix(request.Param("color")))

		io.CopyN(io.Discard, file, 4)

		io.Copy(request.Response(), file)
	})

	app.Help("icons", unlocalized("Ikony"), func(request *Request) template.HTML {
		prefix := app.iconsPrefix
		prefix = strings.TrimRight(prefix, "/")

		icons, err := app.iconsFS.ReadDir(prefix)
		must(err)

		var ret []string
		for _, v := range icons {
			ret = append(ret, v.Name())
		}

		return app.adminTemplates.ExecuteToHTML("help_icons", ret)

	})
}
