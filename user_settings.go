package prago

func (app *App) initUserSettings() {

	/*settingsForm := func(request *Request) *formView {
		user := request.user
		form := newForm()
		formView := form.GetFormView(request)
		formView.Form.Action = "settings"

		name := formView.AddTextInput("name", "")
		name.NameHuman = messages.Get(user.Locale, "Name")
		name.Value = user.Name

		sel := formView.AddSelect("locale", messages.Get(user.Locale, "admin_locale"), availableLocales)
		sel.Value = user.Locale

		formView.AddSubmit("_submit", messages.Get(user.Locale, "admin_edit"))
		return formView
	}*/

	app.Action("settings").Permission(loggedPermission).Name(messages.GetNameFunction("admin_settings")).userMenu().Template("admin_form").DataSource(
		func(request *Request) interface{} {
			user := request.user
			form := newForm()
			form.AJAX = true
			form.Action = "/admin/settings"
			formView := form.GetFormView(request)
			formView.Classes = append(formView.Classes, "prago_form")
			formView.Form.Action = "settings"

			name := formView.AddTextInput("name", "")
			name.NameHuman = messages.Get(user.Locale, "Name")
			name.Value = user.Name

			sel := formView.AddSelect("locale", messages.Get(user.Locale, "admin_locale"), availableLocales)
			sel.Value = user.Locale

			formView.AddSubmit("_submit", messages.Get(user.Locale, "admin_edit"))
			formView.AddCSRFToken(request)
			return formView
			//return settingsForm(request).AddCSRFToken(request)
		},
	)

	app.Action("settings").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			request.RenderJSON(validateSettings(request))

			/*validateCSRF(request)
			u := *request.user
			form := settingsForm(request).AddCSRFToken(request)
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
			}*/
		})

	changePasswordForm := func(request *Request) *formView {
		locale := request.user.Locale
		oldValidator := newValidator(func(field *formItemView) bool {
			if !request.user.isPassword(field.Value) {
				return false
			} else {
				return true
			}
		}, messages.Get(locale, "admin_password_wrong"))

		form := newForm()
		formView := form.GetFormView(request)
		formView.AddPasswordInput("oldpassword",
			messages.Get(locale, "admin_password_old"),
			oldValidator,
		)
		formView.AddPasswordInput("newpassword",
			messages.Get(locale, "admin_password_new"),
			minLengthValidator(messages.Get(locale, "admin_password_length"), 7),
		)
		formView.AddSubmit("_submit", messages.Get(locale, "admin_save"))
		return formView
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

func validateSettings(request *Request) *FormValidation {
	ret := NewFormValidation()
	validateCSRF(request)
	locale := request.user.Locale

	valid := true
	name := request.Params().Get("name")
	if name == "" {
		valid = false
		ret.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
	}

	newLocale := request.Params().Get("locale")
	foundLocale := false
	for _, v := range availableLocales {
		if v[0] == newLocale {
			foundLocale = true
		}
	}
	if !foundLocale {
		valid = false
		ret.AddItemError("locale", "wrong locale")
	}

	if valid {
		u := *request.user
		u.Name = name
		u.Locale = newLocale
		must(request.app.Save(&u))

		request.AddFlashMessage(messages.Get(request.user.Locale, "admin_settings_changed"))
		ret.RedirectionLocaliton = request.app.getAdminURL("")
	}

	return ret
}
