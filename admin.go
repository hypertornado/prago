package prago

import (
	"errors"

	//embed
	_ "embed"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

const adminPathPrefix = "/admin"

func (app *App) initAdminActions() {

	app.accessController.addBeforeAction(func(request *Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")
		request.SetData("locale", localeFromRequest(request))
		request.SetData("admin_header_prefix", adminPathPrefix)
		request.SetData("javascripts", app.javascripts)
		request.SetData("css", app.css)
	})

	app.initSessions()

	googleAPIKey := app.ConfigurationGetStringWithFallback("google", "")
	app.adminController.addAroundAction(func(request *Request, next func()) {
		if request.user == nil {
			request.Redirect(app.getAdminURL("user/login"))
		}

		request.SetData("_csrfToken", app.generateCSRFToken(request.user))
		request.SetData("currentuser", request.user)
		request.SetData("locale", request.user.Locale)
		request.SetData("gravatar", request.user.gravatarURL())

		if !request.user.IsAdmin && !request.user.emailConfirmed() {
			addCurrentFlashMessage(request, messages.Get(request.user.Locale, "admin_flash_not_confirmed"))
		}

		if !request.user.IsAdmin {
			var sysadmin User
			err := app.Query().WhereIs("IsSysadmin", true).Get(&sysadmin)
			var sysadminEmail string
			if err == nil {
				sysadminEmail = sysadmin.Email
			}

			addCurrentFlashMessage(request, messages.Get(request.user.Locale, "admin_flash_not_approved", sysadminEmail))
		}

		request.SetData("main_menu", app.getMainMenu(request))

		request.SetData("google", googleAPIKey)

		next()
	})

	app.Action("markdown").Name(Unlocalized("Nápověda markdown")).hiddenMenu().Template("admin_help_markdown").IsWide()
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

//AddJavascript adds javascript
func (app *App) AddJavascript(url string) {
	app.javascripts = append(app.javascripts, url)
}

//AddCSS adds CSS
func (app *App) AddCSS(url string) {
	app.css = append(app.css, url)
}

func addCurrentFlashMessage(request *Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

func render403(request *Request) {
	request.SetData("message", messages.Get(request.user.Locale, "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request *Request) {
	request.SetData("message", messages.Get(request.user.Locale, "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}
