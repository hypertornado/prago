package prago

import (
	"errors"
	"net/url"
	"time"
)

func initUserRenew(resource *Resource) {

	app := resource.app

	forgottenPasswordForm := func(locale string) *form {
		form := newForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		form.AddSubmit("send", messages.Get(locale, "admin_forgotten_submit"))
		return form
	}

	renderForgot := func(request Request, form *form, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	app.accessController.get(resource.getURL("forgot"), func(request Request) {
		locale := getLocale(request)
		renderForgot(request, forgottenPasswordForm(locale), locale)
	})

	app.accessController.post(resource.getURL("forgot"), func(request Request) {
		email := fixEmail(request.Params().Get("email"))

		var reason = ""
		var user User

		err := app.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if user.emailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
					user.EmailRenewedAt = time.Now()
					err = app.Save(&user)
					if err == nil {
						err = user.sendRenew(request, app)
						if err == nil {
							request.AddFlashMessage(messages.Get(user.Locale, "admin_forgoten_sent", user.Email))
							request.Redirect(app.GetAdminURL("/user/login"))
							return
						}
						reason = "can't send renew email"
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

		request.AddFlashMessage(messages.Get(user.Locale, "admin_forgoten_error", user.Email) + " (" + reason + ")")
		request.Redirect(app.GetAdminURL("user/forgot"))
	})

	renewPasswordForm := func(locale string) (form *form) {
		form = newForm()
		form.Method = "POST"

		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password_new"),
			minLengthValidator(messages.Get(locale, "admin_password_length"), 7))
		passwordInput.Focused = true
		form.AddSubmit("send", messages.Get(locale, "admin_forgoten_set"))
		return
	}

	renderRenew := func(request Request, form *form, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	app.accessController.get(resource.getURL("renew_password"), func(request Request) {
		locale := getLocale(request)
		form := renewPasswordForm(locale)
		renderRenew(request, form, locale)
	})

	app.accessController.post(resource.getURL("renew_password"), func(request Request) {
		locale := getLocale(request)

		form := renewPasswordForm(locale)

		form.BindData(request.Params())
		form.Validate()

		email := request.Params().Get("email")
		email = fixEmail(email)
		token := request.Params().Get("token")

		errStr := messages.Get(locale, "admin_error")

		var user User
		err := app.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if token == user.emailToken(app) {
				if form.Valid {
					err = user.newPassword(request.Params().Get("password"))
					if err == nil {
						err = app.Save(&user)
						if err == nil {
							request.AddFlashMessage(messages.Get(locale, "admin_password_changed"))
							request.Redirect(app.GetAdminURL("user/login"))
							return
						}
					}
				}
			}
		}
		request.AddFlashMessage(errStr)
		request.Redirect(app.GetAdminURL("user/login"))
	})

}

func (user User) getRenewURL(request Request, app *App) string {
	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(app))
	return app.ConfigurationGetString("baseUrl") + app.GetAdminURL("/user/renew_password") + "?" + urlValues.Encode()
}

func (user User) sendRenew(request Request, app *App) error {
	if app.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}

	subject := messages.Get(user.Locale, "admin_forgotten_email_subject", app.name(user.Locale))
	link := user.getRenewURL(request, app)
	body := messages.Get(user.Locale, "admin_forgotten_email_body", link, link, app.name(user.Locale))

	return app.SendEmail(
		user.Name,
		user.Email,
		subject,
		body,
		body,
	)
}
