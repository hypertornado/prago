package prago

import "github.com/gorilla/sessions"

func (app *App) initUserSettings() {

	settingsForm := func(user User) *form {
		form := newForm()
		form.Method = "POST"
		form.Action = "settings"

		name := form.AddTextInput("name", "")
		name.NameHuman = messages.Get(user.Locale, "Name")
		name.Value = user.Name

		sel := form.AddSelect("locale", messages.Get(user.Locale, "admin_locale"), availableLocales)
		sel.Value = user.Locale

		form.AddSubmit("_submit", messages.Get(user.Locale, "admin_edit"))
		return form
	}

	app.Action("settings").Name(messages.GetNameFunction("admin_settings")).userMenu().Template("admin_form").DataSource(
		func(request Request) interface{} {
			return settingsForm(request.getUser()).AddCSRFToken(request)
		},
	)

	app.Action("settings").Method("POST").Handler(
		func(request Request) {
			validateCSRF(request)
			user := request.getUser()
			form := settingsForm(user).AddCSRFToken(request)
			if form.Validate() {
				userResource, err := app.getResourceByItem(&user)
				if err != nil {
					panic(err)
				}
				must(userResource.bindData(&user, user, request.Params(), form.getFilter()))
				must(app.Save(&user))
				request.AddFlashMessage(messages.Get(getLocale(request), "admin_settings_changed"))
				request.Redirect(app.getAdminURL("settings"))
			} else {
				panic("can't validate settings form")
			}
		})

	changePasswordForm := func(request Request) *form {
		user := request.getUser()
		locale := getLocale(request)
		oldValidator := newValidator(func(field *formItem) bool {
			if !user.isPassword(field.Value) {
				return false
			}
			return true
		}, messages.Get(locale, "admin_password_wrong"))

		form := newForm()
		form.Method = "POST"
		form.AddPasswordInput("oldpassword",
			messages.Get(locale, "admin_password_old"),
			oldValidator,
		)
		form.AddPasswordInput("newpassword",
			messages.Get(locale, "admin_password_new"),
			minLengthValidator(messages.Get(locale, "admin_password_length"), 7),
		)
		form.AddSubmit("_submit", messages.Get(locale, "admin_save"))
		return form
	}

	app.Action("password").Name(messages.GetNameFunction("admin_password_change")).userMenu().Template("admin_form").DataSource(
		func(request Request) interface{} {
			return changePasswordForm(request)
		},
	)

	app.Action("password").Method("POST").Handler(func(request Request) {
		form := changePasswordForm(request)
		form.BindData(request.Params())
		form.Validate()
		if form.Valid {
			password := request.Params().Get("newpassword")
			user := request.getUser()
			must(user.newPassword(password))
			must(app.Save(&user))
			request.AddFlashMessage(messages.Get(getLocale(request), "admin_password_changed"))
			request.Redirect(app.getAdminURL(""))
		} else {
			//TODO: better validation and UI of errors
			for _, v := range form.Items {
				for _, e := range v.Errors {
					request.AddFlashMessage(e)
				}
			}
			request.Redirect(app.getAdminURL("password"))
		}
	})

	app.Action("redirect-to-homepage").Name(messages.GetNameFunction("admin_homepage")).userMenu().Handler(func(request Request) {
		request.Redirect("/")
	})

	app.Action("logout").Name(messages.GetNameFunction("admin_log_out")).userMenu().Handler(func(request Request) {
		validateCSRF(request)
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "user_id")
		session.AddFlash(messages.Get(getLocale(request), "admin_logout_ok"))
		must(session.Save(request.Request(), request.Response()))
		request.Redirect(app.getAdminURL("login"))
	})

}
