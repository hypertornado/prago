package prago

func (app *App) initUserSettings() {

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
		},
	)

	app.Action("settings").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			request.RenderJSON(validateSettings(request))

		})

	app.Action("password").Permission(loggedPermission).Name(messages.GetNameFunction("admin_password_change")).userMenu().Template("admin_form").DataSource(
		func(request *Request) interface{} {
			locale := request.user.Locale
			form := newForm()
			form.AJAX = true
			form.Action = "/admin/password"
			formView := form.GetFormView(request)
			formView.Classes = append(formView.Classes, "prago_form")
			formView.AddPasswordInput("oldpassword", messages.Get(locale, "admin_password_old")).Focused = true
			formView.AddPasswordInput("newpassword", messages.Get(locale, "admin_password_new"))
			formView.AddCSRFToken(request)
			formView.AddSubmit("_submit", messages.Get(locale, "admin_save"))
			return formView
		},
	)

	app.Action("password").Permission(loggedPermission).Method("POST").Handler(func(request *Request) {
		request.RenderJSON(validateChangePassword(request))
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

func validateChangePassword(request *Request) *FormValidation {
	ret := NewFormValidation()
	validateCSRF(request)
	locale := request.user.Locale

	valid := true
	oldpassword := request.Params().Get("oldpassword")
	if !request.user.isPassword(oldpassword) {
		valid = false
		ret.AddItemError("oldpassword", messages.Get(locale, "admin_register_password"))
	}

	newpassword := request.Params().Get("newpassword")
	if len(newpassword) < 7 {
		valid = false
		ret.AddItemError("newpassword", messages.Get(locale, "admin_password_length"))
	}

	if valid {
		request.AddFlashMessage(messages.Get(request.user.Locale, "admin_password_changed"))
		ret.RedirectionLocaliton = "/admin"
	}

	return ret
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
