package prago

func initUserSettings(app *App) {

	ActionForm(app, "settings", func(form *Form, request *Request) {
		user := request.getUser()
		form.Title = messages.Get(request.Locale(), "admin_settings")

		name := form.AddTextInput("name", "")
		name.Name = messages.Get(request.Locale(), "Name")
		name.Value = user.Name

		sel := form.AddSelect("locale", messages.Get(request.Locale(), "admin_locale"), availableLocales)
		sel.Value = request.Locale()

		for _, v := range app.settings.settingsArray {
			if request.Authorize(v.permission) {
				val, err := app.getSetting(v.id)
				if err == nil {
					input := form.AddTextInput("setting_"+v.id, v.name(request.Locale()))
					input.Value = val
				} else {
					app.Log().Errorf("can't load setting value '%s': %s", v.id, err)
				}
			}
		}

		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(vc FormValidation, request *Request) {
		locale := request.Locale()
		name := request.Param("name")
		if name == "" {
			vc.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
		}

		newLocale := request.Param("locale")
		foundLocale := false
		for _, v := range availableLocales {
			if v[0] == newLocale {
				foundLocale = true
			}
		}
		if !foundLocale {
			vc.AddItemError("locale", "wrong locale")
		}

		if vc.Valid() {
			user := request.getUser()
			user.Name = name
			user.Locale = newLocale
			must(UpdateItem(app, user))

			for _, v := range app.settings.settingsArray {
				if request.Authorize(v.permission) {
					val := request.Params().Get("setting_" + v.id)
					err := app.saveSetting(v.id, val, request)
					must(err)
					if v.changeCallback != nil {
						v.changeCallback()
					}
				}
			}

			app := request.app
			app.userDataCacheDelete(user.ID)

			request.AddFlashMessage(messages.Get(newLocale, "admin_settings_changed"))
			vc.Redirect("/admin")
		}
	}).Icon("glyphicons-basic-5-settings.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_settings")).userMenu()

	ActionForm(
		app,
		"password",
		func(form *Form, request *Request) {
			locale := request.Locale()
			form.Title = messages.Get(request.Locale(), "admin_password_change")
			oldPassword := form.AddPasswordInput("oldpassword", messages.Get(locale, "admin_password_old"))
			oldPassword.Focused = true
			oldPassword.Autocomplete = "old-password"

			newPassword := form.AddPasswordInput("newpassword", messages.Get(locale, "admin_password_new"))
			newPassword.Focused = true
			newPassword.Autocomplete = "new-password"

			form.AddSubmit(messages.Get(locale, "admin_save"))
		}, func(vc FormValidation, request *Request) {
			locale := request.Locale()

			valid := true
			oldpassword := request.Param("oldpassword")
			user := request.getUser()
			if !user.isPassword(oldpassword) {
				valid = false
				vc.AddItemError("oldpassword", messages.Get(locale, "admin_register_password"))
			}

			newpassword := request.Param("newpassword")
			if len(newpassword) < 7 {
				valid = false
				vc.AddItemError("newpassword", messages.Get(locale, "admin_password_length"))
			}

			if valid {
				request.app.userDataCacheDelete(user.ID)
				request.AddFlashMessage(messages.Get(request.Locale(), "admin_password_changed"))
				vc.Redirect("/admin")
			}
		}).Icon("glyphicons-basic-45-key.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_password_change")).userMenu()

	ActionPlain(app, "redirect-to-homepage", func(request *Request) {
		request.Redirect("/")
	}).Icon("glyphicons-basic-21-home.svg").Permission(loggedPermission).Name(messages.GetNameFunction("boardpage")).userMenu()

	ActionPlain(app, "logout", func(request *Request) {
		validateCSRF(request)
		request.logOutUser()
		request.Redirect(app.getAdminURL("login"))
	}).Icon("glyphicons-basic-432-log-out.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_log_out")).userMenu()
}
