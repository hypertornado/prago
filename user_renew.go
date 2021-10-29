package prago

import (
	"net/url"
	"time"
)

func initUserRenew(resource *Resource) {

	app := resource.app

	/*forgottenPasswordForm := func(request *Request) *formView {
		locale := localeFromRequest(request)
		form := newForm()
		formView := form.GetFormView(request)
		formView.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		formView.AddSubmit("send", messages.Get(locale, "admin_forgotten_submit"))
		return formView
	}

	renderForgot := func(request *Request, form *formView, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}*/

	app.accessController.get(resource.getURL("forgot"), func(request *Request) {
		locale := localeFromRequest(request)

		form := newForm()
		form.AJAX = true
		form.Action = "/admin/user/forgot"
		formView := form.GetFormView(request)
		formView.Classes = append(formView.Classes, "prago_form")
		formView.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		formView.AddSubmit("send", messages.Get(locale, "admin_forgotten_submit"))

		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     formView,
		})

		//renderForgot(request, forgottenPasswordForm(request), locale)
	})

	app.accessController.post(resource.getURL("forgot"), func(request *Request) {
		request.RenderJSON(forgotValidation(request))
	})

	/*app.accessController.post(resource.getURL("forgot_OLD"), func(request *Request) {
		email := fixEmail(request.Params().Get("email"))

		var reason = ""
		var user user

		err := app.Is("email", email).Get(&user)
		if err == nil {
			if user.emailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
					user.EmailRenewedAt = time.Now()
					err = app.Save(&user)
					if err == nil {
						err = user.sendRenew(request, app)
						if err == nil {
							request.AddFlashMessage(messages.Get(user.Locale, "admin_forgoten_sent", user.Email))
							request.Redirect(app.getAdminURL("/user/login"))
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
		request.Redirect(app.getAdminURL("user/forgot"))
	})*/

	/*renewPasswordForm := func(request *Request) *formView {
		locale := localeFromRequest(request)
		form := newForm()
		formView := form.GetFormView(request)

		passwordInput := formView.AddPasswordInput("password", messages.Get(locale, "admin_password_new"),
			minLengthValidator(messages.Get(locale, "admin_password_length"), 7))
		passwordInput.Focused = true
		formView.AddSubmit("send", messages.Get(locale, "admin_forgoten_set"))
		return formView
	}*/

	/*renderRenew := func(request *Request, form *formView, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}*/

	app.accessController.get(resource.getURL("renew_password"), func(request *Request) {
		locale := localeFromRequest(request)
		//form := renewPasswordForm(request)

		form := newForm()
		form.AJAX = true
		form.Action = "/admin/user/renew_password"
		formView := form.GetFormView(request)
		formView.Classes = append(formView.Classes, "prago_form")
		passwordInput := formView.AddPasswordInput("password", messages.Get(locale, "admin_password_new"),
			minLengthValidator(messages.Get(locale, "admin_password_length"), 7))
		passwordInput.Focused = true

		formView.AddHidden("email").Value = request.Params().Get("email")
		formView.AddHidden("token").Value = request.Params().Get("token")

		formView.AddSubmit("send", messages.Get(locale, "admin_forgoten_set"))

		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     formView,
		})

		//renderRenew(request, form, locale)
	})

	app.accessController.post(resource.getURL("renew_password"), func(request *Request) {
		request.RenderJSON(renewPasswordValidation(request))

	})

	/*app.accessController.post(resource.getURL("renew_password_OLD"), func(request *Request) {
		locale := localeFromRequest(request)

		form := renewPasswordForm(request)

		form.BindData(request.Params())
		form.Validate()

		email := request.Params().Get("email")
		email = fixEmail(email)
		token := request.Params().Get("token")

		errStr := messages.Get(locale, "admin_error")

		var user user
		err := app.Is("email", email).Get(&user)
		if err == nil {
			if token == user.emailToken(app) {
				if form.Valid {
					err = user.newPassword(request.Params().Get("password"))
					if err == nil {
						err = app.Save(&user)
						if err == nil {
							request.AddFlashMessage(messages.Get(locale, "admin_password_changed"))
							request.Redirect(app.getAdminURL("user/login"))
							return
						}
					}
				}
			}
		}
		request.AddFlashMessage(errStr)
		request.Redirect(app.getAdminURL("user/login"))
	})*/

}

func renewPasswordValidation(request *Request) *FormValidation {
	ret := NewFormValidation()
	locale := localeFromRequest(request)

	email := request.Params().Get("email")
	email = fixEmail(email)
	token := request.Params().Get("token")

	errStr := messages.Get(locale, "admin_error")

	var user user
	err := request.app.Is("email", email).Get(&user)
	if err == nil {
		if token == user.emailToken(request.app) {
			password := request.Params().Get("password")
			if len(password) >= 7 {
				err = user.newPassword(password)
				if err == nil {
					err = request.app.Save(&user)
					if err == nil {
						request.AddFlashMessage(messages.Get(locale, "admin_password_changed"))
						ret.RedirectionLocaliton = request.app.getAdminURL("user/login") + "?email=" + url.QueryEscape(user.Email)
						return ret
					}
				}
			} else {
				ret.AddItemError("password", messages.Get(locale, "admin_register_password"))
				return ret
			}
		}
	}

	request.AddFlashMessage(errStr)
	ret.RedirectionLocaliton = request.app.getAdminURL("user/login")
	return ret
}

func forgotValidation(request *Request) *FormValidation {
	ret := NewFormValidation()
	app := request.app
	locale := localeFromRequest(request)

	email := fixEmail(request.Params().Get("email"))

	var reason = ""
	var user user
	err := app.Is("email", email).Get(&user)
	if err == nil {
		if user.emailConfirmed() {
			if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
				user.EmailRenewedAt = time.Now()
				err = app.Save(&user)
				if err == nil {
					err = user.sendRenew(request, app)
					if err == nil {
						request.AddFlashMessage(messages.Get(user.Locale, "admin_forgoten_sent", user.Email))
						ret.RedirectionLocaliton = app.getAdminURL("/user/login") + "?email=" + url.QueryEscape(user.Email)
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
		ret.AddError(messages.Get(locale, "admin_forgoten_error", user.Email) + " (" + reason + ")")
	}
	return ret

}

func (user user) getRenewURL(request *Request, app *App) string {
	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(app))
	return app.ConfigurationGetString("baseUrl") + app.getAdminURL("user/renew_password") + "?" + urlValues.Encode()
}

func (user user) sendRenew(request *Request, app *App) error {
	subject := messages.Get(user.Locale, "admin_forgotten_email_subject", app.name(user.Locale))
	link := user.getRenewURL(request, app)
	body := messages.Get(user.Locale, "admin_forgotten_email_body", link, link, app.name(user.Locale))

	return app.Email().To(user.Name, user.Email).Subject(subject).TextContent(body).Send()
}
