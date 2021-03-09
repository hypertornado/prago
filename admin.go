package prago

import (
	"errors"

	_ "embed"

	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/messages"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

const adminPathPrefix = "/admin"

func (app *App) initAdminActions() {

	app.accessController.AddBeforeAction(func(request Request) {
		request.Response().Header().Set("X-XSS-Protection", "1; mode=block")
		request.SetData("locale", getLocale(request))
		request.SetData("admin_header_prefix", adminPathPrefix)
		request.SetData("javascripts", app.javascripts)
		request.SetData("css", app.css)
	})

	app.accessController.AddAroundAction(
		app.createSessionAroundAction(
			app.Config.GetString("random"),
		),
	)

	googleAPIKey := app.Config.GetStringWithFallback("google", "")
	app.AdminController.AddBeforeAction(func(request Request) {
		request.SetData("google", googleAPIKey)
	})

	app.AdminController.AddBeforeAction(func(request Request) {
		session := request.GetData("session").(*sessions.Session)
		userID, ok := session.Values["user_id"].(int64)

		if !ok {
			request.Redirect(app.GetAdminURL("user/login"))
			return
		}

		var user User
		err := app.Query().WhereIs("id", userID).Get(&user)
		if err != nil {
			request.Redirect(app.GetAdminURL("user/login"))
			return
		}

		randomness := app.Config.GetString("random")
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", user.Locale)
		request.SetData("gravatar", user.gravatarURL())

		if !user.IsAdmin && !user.emailConfirmed() {
			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_confirmed"))
		}

		if !user.IsAdmin {
			var sysadmin User
			err := app.Query().WhereIs("IsSysadmin", true).Get(&sysadmin)
			var sysadminEmail string
			if err == nil {
				sysadminEmail = sysadmin.Email
			}

			addCurrentFlashMessage(request, messages.Messages.Get(user.Locale, "admin_flash_not_approved", sysadminEmail))
		}

		request.SetData("main_menu", app.getMainMenu(request))
	})

	app.AdminController.Get(app.GetAdminURL(""), func(request Request) {
		renderNavigationPage(request, adminNavigationPage{
			PageTemplate: "admin_home_navigation",
			PageData:     app.getHomeData(request),
		})
	})

	app.AdminController.Get(app.GetAdminURL("_help/markdown"), func(request Request) {
		request.SetData("admin_yield", "admin_help_markdown")
		request.RenderView("admin_layout")
	})

}

func (app *App) initAdminNotFoundAction() {
	app.AdminController.Get(app.GetAdminURL("*"), render404)
}

//GetAdminURL gets url
func (app App) GetAdminURL(suffix string) string {
	ret := adminPathPrefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//AddAction adds action
func (app *App) AddAction(action Action) {
	bindAction(app, nil, action, false)
	app.rootActions = append(app.rootActions, action)
}

//AddJavascript adds javascript
func (app *App) AddJavascript(url string) {
	app.javascripts = append(app.javascripts, url)
}

//AddCSS adds CSS
func (app *App) AddCSS(url string) {
	app.css = append(app.css, url)
}

//AddFlashMessage adds flash message to request
func AddFlashMessage(request Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	must(session.Save(request.Request(), request.Response()))
}

func addCurrentFlashMessage(request Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

func render403(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}
