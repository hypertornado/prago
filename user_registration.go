package prago

import (
	"errors"
	"fmt"
	"net/url"
	"time"
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
				var user user
				app.Is("email", field.Value).Get(&user)
				return user.Email != field.Value
			}, messages.Get(locale, "admin_email_already_registered")),
		)
		form.AddPasswordInput("password", messages.Get(locale, "admin_register_password"),
			minLengthValidator("", 7),
		)
		form.AddCAPTCHAInput("captcha", "4 + 5 =", valueValidator("9", "Špatná hodnota"))
		form.AddSubmit("send", messages.Get(locale, "admin_register"))
		return form
	}

	renderRegistration := func(request *Request, form *form, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          app,
			Navigation:   app.getNologinNavigation(locale, "registration"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	app.accessController.get(resource.getURL("registration"), func(request *Request) {
		locale := localeFromRequest(request)
		renderRegistration(request, newUserForm(locale), locale)
	})

	app.accessController.post(resource.getURL("registration"), func(request *Request) {
		locale := localeFromRequest(request)
		form := newUserForm(locale)

		form.BindData(request.Params())

		if form.Validate() {
			email := request.Params().Get("email")
			email = fixEmail(email)
			u := &user{}
			u.Email = email
			u.Name = request.Params().Get("name")
			u.IsActive = true
			u.Locale = locale
			must(u.newPassword(request.Params().Get("password")))
			err := u.sendConfirmEmail(request, app)
			if err != nil {
				app.Log().Println(err)
			}
			err = u.sendAdminEmail(request, app)
			if err != nil {
				app.Log().Println(err)
			}

			count, err := app.Query().Count(&user{})
			if err == nil && count == 0 {
				u.Role = sysadminRoleName
			}

			must(app.Create(u))

			request.AddFlashMessage(messages.Get(locale, "admin_confirm_email_send", u.Email))
			request.Redirect(app.getAdminURL("user/login"))
		} else {
			form.GetItemByName("password").Value = ""
			renderRegistration(request, form, locale)
		}
	})

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
