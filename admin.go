package prago

import (
	"errors"
)

// ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

const adminPathPrefix = "/admin"

func (app *App) initAdminActions() {

	app.accessController.addBeforeAction(func(request *Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")
		request.SetData("locale", localeFromRequest(request))
		request.SetData("admin_header_prefix", adminPathPrefix)
	})

	app.adminController.addAroundAction(func(request *Request, next func()) {
		if request.user == nil || !request.user.IsActive {
			request.Redirect(app.getAdminURL("user/login"))
			return
		}

		request.SetData("javascripts", app.javascripts)

		request.SetData("currentuser", request.user)
		request.SetData("locale", request.user.Locale)

		if request.user.Role == "" && !request.user.emailConfirmed() {
			addCurrentFlashMessage(request, messages.Get(request.user.Locale, "admin_flash_not_confirmed"))
		}

		if request.user.Role == "" {
			addCurrentFlashMessage(request, messages.Get(request.user.Locale, "admin_flash_not_approved"))
		}

		request.SetData("main_menu", app.getMainMenu(request))

		next()
	})
	app.Action("markdown").Name(unlocalized("Nápověda markdown")).Permission(loggedPermission).hiddenInMainMenu().Template("admin_help_markdown")

	app.accessController.get("/admin/logo", func(request *Request) {
		request.w.Write(app.logo)
	})
}

func (app *App) initAdminNotFoundAction() {
	app.adminController.get(app.getAdminURL("*"), render404)
}

func (app App) getAdminURL(suffix string) string {
	ret := adminPathPrefix
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

func addCurrentFlashMessage(request *Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

func render403(request *Request) {
	title := messages.Get(request.user.Locale, "admin_403")
	renderPage(request, page{
		Name:         title,
		PageTemplate: "admin_message",
		PageData: map[string]interface{}{
			"message": title,
		},
		HTTPCode: 403,
	})
}

func render404(request *Request) {
	title := messages.Get(request.user.Locale, "admin_404")
	renderPage(request, page{
		Name:         title,
		PageTemplate: "admin_message",
		PageData: map[string]interface{}{
			"message": title,
		},
		HTTPCode: 404,
	})
}
