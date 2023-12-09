package prago

import (
	"errors"
	"fmt"
	"net/http"
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

	app.Action("markdown").Name(unlocalized("Nápověda markdown")).Permission(loggedPermission).Board(nil).View("admin_help_markdown", nil)

	app.accessController.routeHandler("GET", "/admin/logo", func(request *Request) {
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
	app.adminController.routeHandler("GET", app.getAdminURL("*"), func(request *Request) {
		renderErrorPage(request, 404)
	})
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

func renderErrorPage(request *Request, httpCode int) {
	name := messages.Get(request.Locale(), fmt.Sprintf("admin_%d", httpCode))

	if name == "" {
		name = http.StatusText(httpCode)
	}

	pageData := createPageData(request)
	pageData.Messages = append(pageData.Messages, pageMessage{
		Name: name,
	})
	pageData.Name = name
	pageData.HTTPCode = httpCode
	pageData.renderPage(request)
}
