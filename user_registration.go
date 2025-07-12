package prago

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func initUserRegistration(app *App) {

	app.accessController.routeHandler("GET", app.UsersResource.getURL("confirm_email"), func(request *Request) {
		email := request.Param("email")
		token := request.Param("token")

		user := Query[user](app).Is("email", email).First()
		if user != nil {
			if !user.emailConfirmed() {
				if token == user.emailToken(app) {
					user.EmailConfirmedAt = time.Now()
					err := UpdateItem(app, user)
					if err == nil {
						request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_ok"))
						request.Redirect("/admin/user/login")
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
		request.Redirect("/admin/user/login")
	})

	app.nologinFormAction("registration", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		nameInput := form.AddTextInput("name", messages.Get(locale, "Name"))
		nameInput.Focused = true

		emailInput := form.AddEmailInput("email", messages.Get(locale, "admin_email"))
		emailInput.InputMode = "email"
		emailInput.Autocomplete = "email"

		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_register_password"))
		passwordInput.Description = messages.Get(locale, "admin_register_password_description")
		passwordInput.Autocomplete = "new-password"

		form.AddCAPTCHAInput("captcha", "4 + 5 =")
		form.AddSubmit(messages.Get(locale, "admin_register"))
	}, registrationValidation)

}

func registrationValidation(vc FormValidation, request *Request) {
	locale := request.Locale()
	app := request.app

	name := request.Param("name")
	if name == "" {
		vc.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
	}

	email := request.Param("email")
	email = fixEmail(email)
	if !isEmailValid(email) {
		vc.AddItemError("email", messages.Get(locale, "admin_email_not_valid"))
	} else {
		user := Query[user](app).Is("email", email).First()
		if user != nil && user.Email == email {
			vc.AddItemError("email", messages.Get(locale, "admin_email_already_registered"))
		}
	}

	password := request.Param("password")
	if len(password) < 7 {
		vc.AddItemError("password", messages.Get(locale, "admin_register_password"))
	}

	captcha := request.Param("captcha")
	captcha = strings.Trim(captcha, " ")
	if captcha != "9" {
		vc.AddItemError("captcha", messages.Get(locale, "admin_error"))
	}

	if vc.Valid() {
		u := &user{}
		u.Email = email
		u.Name = request.Param("name")
		u.Locale = locale
		must(u.newPassword(request.Param("password")))

		count, err := Query[user](app).Count()
		if err == nil && count == 0 {
			u.Role = sysadminRoleName
		}

		must(CreateItemWithContext(request.Request().Context(), app, u))

		err = u.sendConfirmEmail(app, locale)
		if err != nil {
			app.Log().Println(err)
		}
		err = u.sendAdminEmail(app)
		if err != nil {
			app.Log().Println(err)
		}

		request.AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", u.Email))
		vc.Redirect(app.getAdminURL("user/login") + "?email=" + url.QueryEscape(email))
	}
}

func (u user) sendConfirmEmail(app *App, locale string) error {
	if u.emailConfirmed() {
		return errors.New("email already confirmed")
	}
	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.emailToken(app))

	return app.Mailing(locale, func(md *MailingData) {
		md.AddRecipient(u.Name, u.Email)

		link := app.mustGetSetting("base_url") + app.getAdminURL("user/confirm_email") + "?" + urlValues.Encode()

		md.Name = messages.Get(locale, "admin_confirm_email_subject")

		md.Description = template.HTML(messages.Get(locale, "admin_confirm_email_body"))

		md.Button = &Button{
			Name: messages.Get(locale, "admin_confirm_button"),
			URL:  link,
		}

		md.FooterDescription = template.HTML(fmt.Sprintf("<a href=\"%s\">%s</a>", md.BaseURL, md.AppName))

	})
}

func (u user) sendAdminEmail(app *App) error {
	users := Query[user](app).Is("role", "sysadmin").List()
	for _, receiver := range users {

		err := app.Mailing(receiver.Locale, func(md *MailingData) {
			md.AddRecipient(receiver.Name, receiver.Email)
			md.Name = "New registration"

			md.AddSection("User's name", u.Name)
			md.AddSection("User's email", u.Email)

			md.Button = &Button{
				Name: "Detail",
				URL:  app.BaseURL() + fmt.Sprintf("/admin/user/%d", u.ID),
			}
		})
		if err != nil {
			return err
		}

	}
	return nil
}

const emailRegexpStr = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"

var emailRegexp = regexp.MustCompile(emailRegexpStr)

func isEmailValid(email string) bool {
	return emailRegexp.MatchString(email)
}
