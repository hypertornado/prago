package prago

import (
	"errors"
	"net/url"
	"time"

	"github.com/hypertornado/prago/messages"
)

func initUserRenew(resource *Resource) {

	forgottenPasswordForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Messages.Get(locale, "admin_email")).Focused = true
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_forgotten_submit"))
		return form
	}

	renderForgot := func(request Request, form *Form, locale string) {
		renderNavigationPageNoLogin(request, adminNavigationPage{
			Admin:        resource.App,
			Navigation:   resource.App.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	resource.App.accessController.Get(resource.GetURL("forgot"), func(request Request) {
		locale := getLocale(request)
		renderForgot(request, forgottenPasswordForm(locale), locale)
	})

	resource.App.accessController.Post(resource.GetURL("forgot"), func(request Request) {
		email := fixEmail(request.Params().Get("email"))

		var reason = ""
		var user User

		err := resource.App.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if user.emailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
					user.EmailRenewedAt = time.Now()
					err = resource.App.Save(&user)
					if err == nil {
						err = user.sendRenew(request, resource.App)
						if err == nil {
							AddFlashMessage(request, messages.Messages.Get(user.Locale, "admin_forgoten_sent", user.Email))
							request.Redirect(resource.App.GetAdminURL("/user/login"))
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

		AddFlashMessage(request, messages.Messages.Get(user.Locale, "admin_forgoten_error", user.Email)+" ("+reason+")")
		request.Redirect(resource.App.GetAdminURL("user/forgot"))
	})

	renewPasswordForm := func(locale string) (form *Form) {
		form = NewForm()
		form.Method = "POST"

		passwordInput := form.AddPasswordInput("password", messages.Messages.Get(locale, "admin_password_new"),
			MinLengthValidator(messages.Messages.Get(locale, "admin_password_length"), 7))
		passwordInput.Focused = true
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_forgoten_set"))
		return
	}

	renderRenew := func(request Request, form *Form, locale string) {
		renderNavigationPageNoLogin(request, adminNavigationPage{
			Admin:        resource.App,
			Navigation:   resource.App.getNologinNavigation(locale, "forgot"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	resource.App.accessController.Get(resource.GetURL("renew_password"), func(request Request) {
		locale := getLocale(request)
		form := renewPasswordForm(locale)
		renderRenew(request, form, locale)
	})

	resource.App.accessController.Post(resource.GetURL("renew_password"), func(request Request) {
		locale := getLocale(request)

		form := renewPasswordForm(locale)

		form.BindData(request.Params())
		form.Validate()

		email := request.Params().Get("email")
		email = fixEmail(email)
		token := request.Params().Get("token")

		errStr := messages.Messages.Get(locale, "admin_error")

		var user User
		err := resource.App.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if token == user.emailToken(resource.App) {
				if form.Valid {
					err = user.newPassword(request.Params().Get("password"))
					if err == nil {
						err = resource.App.Save(&user)
						if err == nil {
							AddFlashMessage(request, messages.Messages.Get(locale, "admin_password_changed"))
							request.Redirect(resource.App.GetAdminURL("user/login"))
							return
						}
					}
				}
			}
		}
		AddFlashMessage(request, errStr)
		request.Redirect(resource.App.GetAdminURL("user/login"))
		//form.GetItemByName("password").Value = ""
		//renderLogin(request, form, locale)
	})

}

func (user User) getRenewURL(request Request, app *App) string {
	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(app))
	return app.Config.GetString("baseUrl") + app.prefix + "/user/renew_password?" + urlValues.Encode()
}

func (user User) sendRenew(request Request, admin *App) error {
	if admin.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}

	subject := messages.Messages.Get(user.Locale, "admin_forgotten_email_subject", admin.HumanName)
	link := user.getRenewURL(request, admin)
	body := messages.Messages.Get(user.Locale, "admin_forgotten_email_body", link, link, admin.HumanName)

	return admin.SendEmail(
		user.Name,
		user.Email,
		subject,
		body,
		body,
	)
}
