package prago

func initUserSettings(app *App) {

	app.FormAction("settings", func(form *Form, request *Request) {
		user := request.getUser()
		form.Title = messages.Get(request.Locale(), "admin_settings")

		name := form.AddTextInput("name", "")
		name.Name = messages.Get(request.Locale(), "Name")
		name.Value = user.Name

		sel := form.AddSelect("locale", messages.Get(request.Locale(), "admin_locale"), availableLocales)
		sel.Value = request.Locale()

		for _, v := range app.settings.settingsArray {
			if request.Authorize(v.permission) {
				val, err := app.GetSetting(v.id)
				if err == nil {
					input := form.AddTextInput("setting_"+v.id, v.name(request.Locale()))
					input.Value = val
				} else {
					app.Log().Errorf("can't load setting value '%s': %s", v.id, err)
				}
			}
		}

		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(vc ValidationContext) {
		locale := vc.Locale()
		valid := true
		name := vc.GetValue("name")
		if name == "" {
			valid = false
			vc.AddItemError("name", messages.Get(locale, "admin_user_name_not_empty"))
		}

		newLocale := vc.GetValue("locale")
		foundLocale := false
		for _, v := range availableLocales {
			if v[0] == newLocale {
				foundLocale = true
			}
		}
		if !foundLocale {
			valid = false
			vc.AddItemError("locale", "wrong locale")
		}

		if valid {
			user := vc.Request().getUser()
			user.Name = name
			user.Locale = newLocale
			must(UpdateItem(app, user))

			for _, v := range app.settings.settingsArray {
				request := vc.Request()
				if request.Authorize(v.permission) {
					val := request.Params().Get("setting_" + v.id)
					err := app.saveSetting(v.id, val, request)
					must(err)
					if v.changeCallback != nil {
						v.changeCallback()
					}
				}
			}

			app := vc.Request().app
			app.userDataCacheDelete(user.ID)

			vc.Request().AddFlashMessage(messages.Get(newLocale, "admin_settings_changed"))
			vc.Validation().RedirectionLocation = app.getAdminURL("")
		}
	}).Icon("glyphicons-basic-5-settings.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_settings")).userMenu()

	app.FormAction(
		"password",
		func(form *Form, request *Request) {
			locale := request.Locale()
			form.Title = messages.Get(request.Locale(), "admin_password_change")
			form.AddPasswordInput("oldpassword", messages.Get(locale, "admin_password_old")).Focused = true
			form.AddPasswordInput("newpassword", messages.Get(locale, "admin_password_new"))
			form.AddSubmit(messages.Get(locale, "admin_save"))
		}, func(vc ValidationContext) {
			request := vc.Request()
			locale := request.Locale()

			valid := true
			oldpassword := vc.GetValue("oldpassword")
			user := request.getUser()
			if !user.isPassword(oldpassword) {
				valid = false
				vc.AddItemError("oldpassword", messages.Get(locale, "admin_register_password"))
			}

			newpassword := vc.GetValue("newpassword")
			if len(newpassword) < 7 {
				valid = false
				vc.AddItemError("newpassword", messages.Get(locale, "admin_password_length"))
			}

			if valid {
				request.app.userDataCacheDelete(user.ID)
				request.AddFlashMessage(messages.Get(request.Locale(), "admin_password_changed"))
				vc.Validation().RedirectionLocation = "/admin"
			}
		}).Icon("glyphicons-basic-45-key.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_password_change")).userMenu()

	app.Action("redirect-to-homepage").Icon("glyphicons-basic-21-home.svg").Permission(loggedPermission).Name(messages.GetNameFunction("boardpage")).userMenu().Handler(func(request *Request) {
		request.Redirect("/")
	})

	app.Action("logout").Icon("glyphicons-basic-432-log-out.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_log_out")).userMenu().Handler(func(request *Request) {
		validateCSRF(request)
		request.logOutUser()
		request.Redirect(app.getAdminURL("login"))
	})
}
