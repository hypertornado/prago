package prago

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

func initUserRegistration(resource *Resource) {
	app := resource.app

	app.accessController.get(resource.getURL("confirm_email"), func(request Request) {
		email := request.Params().Get("email")
		token := request.Params().Get("token")

		var user User
		err := app.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if !user.emailConfirmed() {
				if token == user.emailToken(app) {
					user.EmailConfirmedAt = time.Now()
					err = app.Save(&user)
					if err == nil {
						request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_ok"))
						request.Redirect(app.GetAdminURL("user/login"))
						return
					}
				}
			}
		}

		request.AddFlashMessage(messages.Get(user.Locale, "admin_confirm_email_fail"))
		request.Redirect(app.GetAdminURL("user/login"))
	})

	newUserForm := func(locale string) *form {
		form := newForm()
		form.Method = "POST"
		form.AddTextInput("name", messages.Get(locale, "Name"),
			nonEmptyValidator(messages.Get(locale, "admin_user_name_not_empty")),
		)
		form.AddEmailInput("email", messages.Get(locale, "admin_email"),
			emailValidator(messages.Get(locale, "admin_email_not_valid")),
			newValidator(func(field *formItem) bool {
				if len(field.Errors) != 0 {
					return true
				}
				var user User
				app.Query().WhereIs("email", field.Value).Get(&user)
				if user.Email == field.Value {
					return false
				}
				return true
			}, messages.Get(locale, "admin_email_already_registered")),
		)
		form.AddPasswordInput("password", messages.Get(locale, "admin_register_password"),
			minLengthValidator("", 7),
		)
		form.AddCAPTCHAInput("captcha", "4 + 5 =", valueValidator("9", "Špatná hodnota"))
		form.AddSubmit("send", messages.Get(locale, "admin_register"))
		return form
	}

	renderRegistration := func(request Request, form *form, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "registration"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	app.accessController.get(resource.getURL("registration"), func(request Request) {
		locale := getLocale(request)
		renderRegistration(request, newUserForm(locale), locale)
	})

	app.accessController.post(resource.getURL("registration"), func(request Request) {
		locale := getLocale(request)
		form := newUserForm(locale)

		form.BindData(request.Params())

		if form.Validate() {
			email := request.Params().Get("email")
			email = fixEmail(email)
			user := &User{}
			user.Email = email
			user.Name = request.Params().Get("name")
			user.IsActive = true
			user.Locale = locale
			must(user.newPassword(request.Params().Get("password")))
			err := user.sendConfirmEmail(request, app)
			if err != nil {
				app.Log().Println(err)
			}
			err = user.sendAdminEmail(request, app)
			if err != nil {
				app.Log().Println(err)
			}

			count, err := app.Query().Count(&User{})
			if err == nil && count == 0 {
				user.IsAdmin = true
				user.Role = "sysadmin"
			}

			must(app.Create(user))

			request.AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", user.Email))
			request.Redirect(app.GetAdminURL("user/login"))
		} else {
			form.GetItemByName("password").Value = ""
			renderRegistration(request, form, locale)
		}
	})

}

func (user User) sendConfirmEmail(request Request, app *App) error {

	if user.emailConfirmed() {
		return errors.New("email already confirmed")
	}

	if app.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}

	locale := getLocale(request)

	urlValues := make(url.Values)
	urlValues.Add("email", user.Email)
	urlValues.Add("token", user.emailToken(app))

	subject := messages.Get(locale, "admin_confirm_email_subject", app.name(user.Locale))
	link := app.ConfigurationGetString("baseUrl") + app.GetAdminURL("/user/confirm_email") + "?" + urlValues.Encode()
	body := messages.Get(locale, "admin_confirm_email_body", link, link, app.name(user.Locale))

	return app.SendEmail(
		user.Name,
		user.Email,
		subject,
		body,
		body,
	)

}

func (user User) sendAdminEmail(request Request, a *App) error {
	if a.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}
	var users []*User
	err := a.Query().WhereIs("issysadmin", true).Get(&users)
	if err != nil {
		return err
	}
	for _, receiver := range users {

		body := fmt.Sprintf("New user registered on %s: %s (%s)", a.name(user.Locale), user.Email, user.Name)

		err = a.SendEmail(
			receiver.Name,
			receiver.Email,
			"New registration on "+a.name(user.Locale),
			body,
			body,
		)

		if err != nil {
			return err
		}
	}
	return nil
}
