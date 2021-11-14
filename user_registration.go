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

		var user user
		err := app.Is("email", email).Get(&user)
		if err == nil {
			if !user.emailConfirmed() {
				if token == user.emailToken(app) {
					user.EmailConfirmedAt = time.Now()
					err = app.Save(&user)
					if err == nil {
						request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_ok"))
						request.Redirect(app.getAdminURL("user/login"))
						return
					}
				}
			}
		}

		request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_fail"))
		request.Redirect(app.getAdminURL("user/login"))
	})

	app.accessController.get(resource.getURL("registration"), func(request *Request) {
		locale := localeFromRequest(request)
		form := NewForm("/admin/user/registration")
		form.AddTextInput("name", messages.Get(locale, "Name")).Focused = true
		form.AddEmailInput("email", messages.Get(locale, "admin_email"))
		form.AddPasswordInput("password", messages.Get(locale, "admin_register_password")).Description = messages.Get(locale, "admin_register_password_description")
		form.AddCAPTCHAInput("captcha", "4 + 5 =")
		form.AddSubmit("send", messages.Get(locale, "admin_register"))

		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "registration"),
			PageTemplate: "admin_form",
			PageData:     form,
		})

	})

	app.accessController.post(resource.getURL("registration"), func(request *Request) {
		request.RenderJSON(registrationValidation(request))
	})
}

func registrationValidation(request *Request) *formValidation {
	ret := NewFormValidation()

	valid := true
	locale := localeFromRequest(request)

	name := request.Params().Get("name")
	if name == "" {
		valid = false
		ret.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
	}

	email := request.Params().Get("email")
	email = fixEmail(email)
	if !govalidator.IsEmail(email) {
		valid = false
		ret.AddItemError("email", messages.Get(locale, "admin_email_not_valid"))
	} else {
		var user user
		request.app.Is("email", email).Get(&user)
		if user.Email == email {
			valid = false
			ret.AddItemError("email", messages.Get(locale, "admin_email_already_registered"))
		}
	}

	password := request.Params().Get("password")
	if len(password) < 7 {
		valid = false
		ret.AddItemError("password", messages.Get(locale, "admin_register_password"))
	}

	captcha := request.Params().Get("captcha")
	captcha = strings.Trim(captcha, " ")
	if captcha != "9" {
		valid = false
		ret.AddItemError("captcha", messages.Get(locale, "admin_error"))
	}

	if valid {
		u := &user{}
		u.Email = email
		u.Name = request.Params().Get("name")
		u.IsActive = true
		u.Locale = locale
		must(u.newPassword(request.Params().Get("password")))
		err := u.sendConfirmEmail(request, request.app)
		if err != nil {
			request.app.Log().Println(err)
		}
		err = u.sendAdminEmail(request, request.app)
		if err != nil {
			request.app.Log().Println(err)
		}

		count, err := request.app.Query().Count(&user{})
		if err == nil && count == 0 {
			u.Role = sysadminRoleName
		}

		must(request.app.Create(u))

		request.AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", u.Email))
		ret.RedirectionLocaliton = request.app.getAdminURL("user/login") + "?email=" + url.QueryEscape(email)

	}

	return ret

}

func (u user) sendConfirmEmail(request *Request, app *App) error {

	if u.emailConfirmed() {
		return errors.New("email already confirmed")
	}

	locale := localeFromRequest(request)
	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.emailToken(app))

	subject := messages.Get(locale, "admin_confirm_email_subject", app.name(u.Locale))
	link := app.ConfigurationGetString("baseUrl") + app.getAdminURL("user/confirm_email") + "?" + urlValues.Encode()
	body := messages.Get(locale, "admin_confirm_email_body", link, link, app.name(u.Locale))
	return app.Email().To(u.Name, u.Email).Subject(subject).HTMLContent(body).Send()
}

func (u user) sendAdminEmail(request *Request, a *App) error {
	var users []*user
	err := a.Query().Is("role", "sysadmin").Get(&users)
	if err != nil {
		return err
	}
	for _, receiver := range users {

		body := fmt.Sprintf("New user registered on %s: %s (%s)", a.name(u.Locale), u.Email, u.Name)

		err = request.app.Email().To(receiver.Name, receiver.Email).Subject("New registration on " + a.name(u.Locale)).HTMLContent(body).Send()
		if err != nil {
			return err
		}
	}
	return nil
}
