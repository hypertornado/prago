package prago

import (
	"context"
	"net/url"
	"time"
)

func initUserRenew(app *App) {

	app.nologinFormAction("forgot", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		form.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		form.AddSubmit(messages.Get(locale, "admin_forgotten_submit"))
	}, func(vc ValidationContext) {
		request := vc.Request()
		ctx := request.r.Context()
		email := fixEmail(request.Param("email"))

		var reason = ""
		user := app.UsersResource.Is(ctx, "email", email).First()
		if user != nil {
			if user.emailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
					user.EmailRenewedAt = time.Now()
					err := app.UsersResource.Update(ctx, user)
					if err == nil {
						err = app.sendRenewPasswordEmail(ctx, *user)
						if err == nil {
							request.AddFlashMessage(messages.Get(user.Locale, "admin_forgoten_sent", user.Email))
							vc.Validation().RedirectionLocaliton = app.getAdminURL("/user/login") + "?email=" + url.QueryEscape(user.Email)
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
			vc.AddError(messages.Get(vc.Locale(), "admin_forgoten_error", user.Email) + " (" + reason + ")")
		}
	})

	app.nologinFormAction("renew_password", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password_new"))
		passwordInput.Focused = true

		form.AddHidden("email").Value = request.Param("email")
		form.AddHidden("token").Value = request.Param("token")
		form.AddSubmit(messages.Get(locale, "admin_forgoten_set"))
	}, func(vc ValidationContext) {
		ctx := vc.Request().r.Context()
		email := vc.GetValue("email")
		email = fixEmail(email)
		token := vc.GetValue("token")

		errStr := messages.Get(vc.Locale(), "admin_error")

		u := app.UsersResource.Is(ctx, "email", email).First()
		if u != nil {
			if token == u.emailToken(vc.Context(), app) {
				password := vc.GetValue("password")
				if len(password) >= 7 {
					err := u.newPassword(password)
					if err == nil {
						err = app.UsersResource.Update(ctx, u)
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

func (app *App) getRenewPasswordURL(ctx context.Context, user user) string {
	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(ctx, app))
	return app.MustGetSetting(ctx, "base_url") + app.getAdminURL("user/renew_password") + "?" + urlValues.Encode()
}

func (app *App) sendRenewPasswordEmail(ctx context.Context, user user) error {
	subject := messages.Get(user.Locale, "admin_forgotten_email_subject", app.name(user.Locale))
	link := app.getRenewPasswordURL(ctx, user)
	body := messages.Get(user.Locale, "admin_forgotten_email_body", link, link, app.name(user.Locale))

	return app.Email().To(user.Name, user.Email).Subject(subject).TextContent(body).Send()
}
