package prago

import (
	"net/url"
	"time"
)

func initUserRenew(resource *Resource) {
	resource.app.nologinFormAction("forgot", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		form.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		form.AddSubmit(messages.Get(locale, "admin_forgotten_submit"))
	}, func(vc ValidationContext) {
		request := vc.Request()
		app := request.app

		email := fixEmail(request.Params().Get("email"))

		var reason = ""
		var u user
		err := app.Is("email", email).Get(&u)
		if err == nil {
			if u.emailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(u.EmailRenewedAt) {
					u.EmailRenewedAt = time.Now()
					err = GetResource[user](resource.app).Update(&u)
					//err = app.Save(&user)
					if err == nil {
						err = app.sendRenewPasswordEmail(u)
						if err == nil {
							request.AddFlashMessage(messages.Get(u.Locale, "admin_forgoten_sent", u.Email))
							vc.Validation().RedirectionLocaliton = app.getAdminURL("/user/login") + "?email=" + url.QueryEscape(u.Email)
						} else {
							reason = "can't send renew email"
						}
					} else {
						reason = "unexpected error"
					}
				} else {
					reason = "email already renewed within last day"
				}
			} else {
				reason = "email not confirmed"
			}
		} else {
			reason = "user not found"
		}

		if reason != "" {
			vc.AddError(messages.Get(vc.Locale(), "admin_forgoten_error", u.Email) + " (" + reason + ")")
		}
	})

	resource.app.nologinFormAction("renew_password", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password_new"))
		passwordInput.Focused = true

		form.AddHidden("email").Value = request.Params().Get("email")
		form.AddHidden("token").Value = request.Params().Get("token")
		form.AddSubmit(messages.Get(locale, "admin_forgoten_set"))
	}, func(vc ValidationContext) {
		app := vc.Request().app

		email := vc.GetValue("email")
		email = fixEmail(email)
		token := vc.GetValue("token")

		errStr := messages.Get(vc.Locale(), "admin_error")

		var u user
		err := app.Is("email", email).Get(&u)
		if err == nil {
			if token == u.emailToken(app) {
				password := vc.GetValue("password")
				if len(password) >= 7 {
					err = u.newPassword(password)
					if err == nil {
						err = GetResource[user](resource.app).Update(&u)
						//err = app.Save(&user)
						if err == nil {
							vc.Request().AddFlashMessage(messages.Get(vc.Locale(), "admin_password_changed"))
							vc.Validation().RedirectionLocaliton = app.getAdminURL("user/login") + "?email=" + url.QueryEscape(u.Email)
							return
						}
					}
				} else {
					vc.AddItemError("password", messages.Get(vc.Locale(), "admin_register_password"))
					return
				}
			}
		}

		vc.Request().AddFlashMessage(errStr)
		vc.Validation().RedirectionLocaliton = app.getAdminURL("user/login")
	})
}

func (app *App) getRenewPasswordURL(user user) string {
	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(app))
	return app.ConfigurationGetString("baseUrl") + app.getAdminURL("user/renew_password") + "?" + urlValues.Encode()
}

func (app *App) sendRenewPasswordEmail(user user) error {
	subject := messages.Get(user.Locale, "admin_forgotten_email_subject", app.name(user.Locale))
	link := app.getRenewPasswordURL(user)
	body := messages.Get(user.Locale, "admin_forgotten_email_body", link, link, app.name(user.Locale))

	return app.Email().To(user.Name, user.Email).Subject(subject).TextContent(body).Send()
}
