package prago

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

func initUserRegistration(resource *Resource) {
	app := resource.app

	app.accessController.get(resource.getURL("confirm_email"), func(request *Request) {
		email := request.Params().Get("email")
		token := request.Params().Get("token")

		var u user
		err := app.Is("email", email).Get(&u)
		if err == nil {
			if !u.emailConfirmed() {
				if token == u.emailToken(app) {
					u.EmailConfirmedAt = time.Now()
					err = GetResource[user](resource.app).Update(&u)
					//err = app.Save(&user)
					if err == nil {
						request.AddFlashMessage(messages.Get(u.Locale, "admin_confirm_email_ok"))
						request.Redirect(app.getAdminURL("user/login"))
						return
					}
				}
			}
		}

		request.AddFlashMessage(messages.Get(u.Locale, "admin_confirm_email_fail"))
		request.Redirect(app.getAdminURL("user/login"))
	})

	resource.app.nologinFormAction("registration", func(form *Form, request *Request) {
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
		var user user
		app.Is("email", email).Get(&user)
		if user.Email == email {
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
		u.IsActive = true
		u.Locale = locale
		must(u.newPassword(vc.GetValue("password")))
		err := u.sendConfirmEmail(app, locale)
		if err != nil {
			app.Log().Println(err)
		}
		err = u.sendAdminEmail(app)
		if err != nil {
			app.Log().Println(err)
		}

		count, err := app.Query().Count(&user{})
		if err == nil && count == 0 {
			u.Role = sysadminRoleName
		}

		must(GetResource[user](app).Create(u))

		vc.Request().AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", u.Email))
		vc.Validation().RedirectionLocaliton = app.getAdminURL("user/login") + "?email=" + url.QueryEscape(email)

	}
}

func (u user) sendConfirmEmail(app *App, locale string) error {
	if u.emailConfirmed() {
		return errors.New("email already confirmed")
	}
	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.emailToken(app))

	subject := messages.Get(locale, "admin_confirm_email_subject", app.name(u.Locale))
	link := app.ConfigurationGetString("baseUrl") + app.getAdminURL("user/confirm_email") + "?" + urlValues.Encode()
	body := messages.Get(locale, "admin_confirm_email_body", link, link, app.name(u.Locale))
	return app.Email().To(u.Name, u.Email).Subject(subject).HTMLContent(body).Send()
}

func (u user) sendAdminEmail(app *App) error {
	var users []*user
	err := app.Query().Is("role", "sysadmin").Get(&users)
	if err != nil {
		return err
	}
	for _, receiver := range users {

		body := fmt.Sprintf("New user registered on %s: %s (%s)", app.name(u.Locale), u.Email, u.Name)

		err = app.Email().To(receiver.Name, receiver.Email).Subject("New registration on " + app.name(u.Locale)).HTMLContent(body).Send()
		if err != nil {
			return err
		}
	}
	return nil
}
