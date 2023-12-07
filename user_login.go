package prago

import (
	"time"
)

func initUserLogin(app *App) {

	app.UsersResource.ItemAction("loginas").Name(unlocalized("Přihlásit se jako")).Permission(sysadminPermission).Handler(
		func(user *user, request *Request) {
			request.logInUser(user)
			request.Redirect(app.getAdminURL(""))
		},
	)

	app.nologinFormAction("login", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		emailValue := request.Param("email")
		emailInput := form.AddEmailInput("email", messages.Get(locale, "admin_email"))
		if emailValue == "" {
			emailInput.Focused = true
		}
		emailInput.Value = request.Param("email")
		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		if emailValue != "" {
			passwordInput.Focused = true
		}
		form.AddSubmit(messages.Get(locale, "admin_login_action"))
	}, func(vc ValidationContext) {
		locale := vc.Locale()
		email := vc.GetValue("email")
		email = fixEmail(email)
		request := vc.Request()
		password := vc.GetValue("password")

		user := Query[user](app).Is("email", email).First()
		if user == nil {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		if !user.isPassword(password) {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		user.LoggedInTime = time.Now()
		user.LoggedInUseragent = request.Request().UserAgent()
		user.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

		must(app.UsersResource.Update(vc.Context(), user))
		request.logInUser(user)
		request.AddFlashMessage(messages.Get(user.Locale, "admin_login_ok"))

		vc.Validation().RedirectionLocaliton = request.app.getAdminURL("")
	})

}
