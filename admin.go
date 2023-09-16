package prago

import (
	"errors"
	"strings"
)

// ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

func (app *App) initAdminActions() {

	app.adminController.addAroundAction(func(request *Request, next func()) {
		if request.UserID() == 0 {
			request.Redirect(app.getAdminURL("user/login"))
			return
		}
		next()
	})

	app.Action("markdown").Name(unlocalized("Nápověda markdown")).Permission(loggedPermission).hiddenInMenu().Template("admin_help_markdown")

	app.accessController.get("/admin/logo", func(request *Request) {
		if app.logo != nil {
			if strings.HasPrefix(string(app.logo), "<svg") {
				request.Response().Header().Add("Content-Type", "image/svg+xml")
			}
			request.w.Write(app.logo)
			return
		}

		iconName := "glyphicons-basic-697-directions-sign.svg"
		if app.icon != "" {
			iconName = app.icon
		}
		iconData, err := app.loadIcon(iconName, "444444")
		must(err)

		request.Response().Header().Add("Content-Type", "image/svg+xml")
		request.w.Write(iconData)
	})
}

func (app *App) initAdminNotFoundAction() {
	app.adminController.get(app.getAdminURL("*"), render404)
}

func (app App) getAdminURL(suffix string) string {
	ret := "/admin"
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

// JavascriptPath adds javascript
func (app *App) JavascriptPath(url string) *App {
	app.javascripts = append(app.javascripts, url)
	return app
}

func render403(request *Request) {
	title := messages.Get(request.Locale(), "admin_403")
	header := BoxHeader{
		Name: title,
	}
	renderPage(request, page{
		Name:         title,
		PageTemplate: "admin_message",
		PageData: map[string]interface{}{
			"box_header": header,
		},
		HTTPCode: 403,
	})
}

func render404(request *Request) {
	title := messages.Get(request.Locale(), "admin_404")
	header := BoxHeader{
		Name: title,
	}

	renderPage(request, page{
		Name:         title,
		PageTemplate: "admin_message",
		PageData: map[string]interface{}{
			"box_header": header,
		},
		HTTPCode: 404,
	})
}
