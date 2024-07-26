package prago

import (
	"strings"
	"time"
)

func initUserLogin(app *App) {
	ActionResourceItemPlain[user](app, "loginas",
		func(user *user, request *Request) {
			request.logInUser(user)
			request.Redirect("/admin")
		}).Name(unlocalized("Přihlásit se jako")).Permission(sysadminPermission)

	app.nologinFormAction("login", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		emailValue := request.Param("email")
		emailInput := form.AddEmailInput("email", messages.Get(locale, "admin_email_or_username"))
		emailInput.InputMode = "email"
		emailInput.Autocomplete = "email"
		if emailValue == "" {
			emailInput.Focused = true
		}
		emailInput.Value = request.Param("email")
		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		passwordInput.Autocomplete = "current-password"
		if emailValue != "" {
			passwordInput.Focused = true
		}

		form.AddHidden("redirect_url").Value = request.Param("redirect")

		form.AddSubmit(messages.Get(locale, "admin_login_action"))
	}, func(vc Validation, request *Request) {
		locale := vc.Locale()
		email := vc.GetValue("email")
		email = fixEmail(email)
		password := vc.GetValue("password")

		q := Query[user](app)
		if email != "" && !strings.Contains(email, "@") {
			q.Is("username", email)
		} else {
			q.Is("email", email)
		}

		user := q.First()
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

		must(UpdateItem(app, user))
		request.logInUser(user)
		request.AddFlashMessage(messages.Get(user.Locale, "admin_login_ok"))

		redirectURL := request.Param("redirect_url")
		if !strings.HasPrefix(redirectURL, "/") || redirectURL == "/admin/login" {
			redirectURL = "/admin"
		}

		vc.Validation().RedirectionLocation = redirectURL
	})

}
