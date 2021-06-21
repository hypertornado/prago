package prago

func (app *App) initUserSettings() {

	settingsForm := func(user *user) *form {
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

	app.Action("settings").Permission(loggedPermission).Name(messages.GetNameFunction("admin_settings")).userMenu().Template("admin_form").DataSource(
		func(request *Request) interface{} {
			return settingsForm(request.user).AddCSRFToken(request)
		},
	)

	app.Action("settings").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			validateCSRF(request)
			u := *request.user
			form := settingsForm(request.user).AddCSRFToken(request)
			if form.Validate() {
				userResource, err := app.getResourceByItem(&u)
				if err != nil {
					panic(err)
				}
				must(userResource.bindData(&u, request.user, request.Params(), nil))
				must(app.Save(&u))
				request.AddFlashMessage(messages.Get(request.user.Locale, "admin_settings_changed"))
				request.Redirect(app.getAdminURL("settings"))
			} else {
				panic("can't validate settings form")
			}
		})

	changePasswordForm := func(request *Request) *form {
		locale := request.user.Locale
		oldValidator := newValidator(func(field *formItem) bool {
			if !request.user.isPassword(field.Value) {
				return false
			} else {
				return true
			}
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

	app.Action("password").Permission(loggedPermission).Name(messages.GetNameFunction("admin_password_change")).userMenu().Template("admin_form").DataSource(
		func(request *Request) interface{} {
			return changePasswordForm(request)
		},
	)

	app.Action("password").Permission(loggedPermission).Method("POST").Handler(func(request *Request) {
		form := changePasswordForm(request)
		form.BindData(request.Params())
		form.Validate()
		if form.Valid {
			password := request.Params().Get("newpassword")
			user := *request.user
			must(user.newPassword(password))
			must(app.Save(&user))
			request.AddFlashMessage(messages.Get(request.user.Locale, "admin_password_changed"))
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

	app.Action("redirect-to-homepage").Permission(loggedPermission).Name(messages.GetNameFunction("admin_homepage")).userMenu().Handler(func(request *Request) {
		request.Redirect("/")
	})

	app.Action("logout").Permission(loggedPermission).Name(messages.GetNameFunction("admin_log_out")).userMenu().Handler(func(request *Request) {
		validateCSRF(request)
		request.logOutUser()
		request.Redirect(app.getAdminURL("login"))
	})

}
