package prago

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

func initUserRegistration(app *App) {

	app.accessController.routeHandler("GET", app.UsersResource.data.getURL("confirm_email"), func(request *Request) {
		ctx := request.r.Context()
		email := request.Param("email")
		token := request.Param("token")

		user := app.UsersResource.Query(ctx).Is("email", email).First()
		if user != nil {
			if !user.emailConfirmed() {
				if token == user.emailToken(request.r.Context(), app) {
					user.EmailConfirmedAt = time.Now()
					err := app.UsersResource.Update(ctx, user)
					if err == nil {
						request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_ok"))
						request.Redirect(app.getAdminURL("user/login"))
						return
					}
				}
			}
		}

		locale := "en"
		if user != nil {
			locale = user.Locale
		}

		request.AddFlashMessage(messages.Get(locale, "admin_confirm_email_fail"))
		request.Redirect(app.getAdminURL("user/login"))
	})

	app.nologinFormAction("registration", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		form.AddTextInput("name", messages.Get(locale, "Name")).Focused = true
		form.AddEmailInput("email", messages.Get(locale, "admin_email"))
		form.AddPasswordInput("password", messages.Get(locale, "admin_register_password")).Description = messages.Get(locale, "admin_register_password_description")
		form.AddCAPTCHAInput("captcha", "4 + 5 =")
		form.AddSubmit(messages.Get(locale, "admin_register"))
	}, registrationValidation)

}

func registrationValidation(vc ValidationContext) {
	valid := true
	locale := vc.Locale()
	app := vc.Request().app

	name := vc.GetValue("name")
	if name == "" {
		valid = false
		vc.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
	}

	email := vc.GetValue("email")
	email = fixEmail(email)
	if !govalidator.IsEmail(email) {
		valid = false
		vc.AddItemError("email", messages.Get(locale, "admin_email_not_valid"))
	} else {
		user := app.UsersResource.Query(vc.Context()).Is("email", email).First()
		if user != nil && user.Email == email {
			valid = false
			vc.AddItemError("email", messages.Get(locale, "admin_email_already_registered"))
		}
	}

	password := vc.GetValue("password")
	if len(password) < 7 {
		valid = false
		vc.AddItemError("password", messages.Get(locale, "admin_register_password"))
	}

	captcha := vc.GetValue("captcha")
	captcha = strings.Trim(captcha, " ")
	if captcha != "9" {
		valid = false
		vc.AddItemError("captcha", messages.Get(locale, "admin_error"))
	}

	if valid {
		u := &user{}
		u.Email = email
		u.Name = vc.GetValue("name")
		u.Locale = locale
		must(u.newPassword(vc.GetValue("password")))
		err := u.sendConfirmEmail(vc.Context(), app, locale)
		if err != nil {
			app.Log().Println(err)
		}
		err = u.sendAdminEmail(vc.Context(), app)
		if err != nil {
			app.Log().Println(err)
		}

		count, err := app.UsersResource.Query(vc.Context()).Count()
		if err == nil && count == 0 {
			u.Role = sysadminRoleName
		}

		must(app.UsersResource.Create(vc.Context(), u))

		vc.Request().AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", u.Email))
		vc.Validation().RedirectionLocaliton = app.getAdminURL("user/login") + "?email=" + url.QueryEscape(email)

	}
}

func (u user) sendConfirmEmail(ctx context.Context, app *App, locale string) error {
	if u.emailConfirmed() {
		return errors.New("email already confirmed")
	}
	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.emailToken(ctx, app))

	subject := messages.Get(locale, "admin_confirm_email_subject", app.name(u.Locale))
	link := app.MustGetSetting(ctx, "base_url") + app.getAdminURL("user/confirm_email") + "?" + urlValues.Encode()
	body := messages.Get(locale, "admin_confirm_email_body", link, link, app.name(u.Locale))
	return app.Email().To(u.Name, u.Email).Subject(subject).HTMLContent(body).Send()
}

func (u user) sendAdminEmail(ctx context.Context, app *App) error {
	res := GetResource[user](app)
	users := res.Query(ctx).Is("role", "sysadmin").List()
	for _, receiver := range users {
		body := fmt.Sprintf("New user registered on %s: %s (%s)", app.name(u.Locale), u.Email, u.Name)

		err := app.Email().To(receiver.Name, receiver.Email).Subject("New registration on " + app.name(u.Locale)).HTMLContent(body).Send()
		if err != nil {
			return err
		}
	}
	return nil
}
