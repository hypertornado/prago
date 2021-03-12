package prago

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

	app.AdminController.Get(app.GetAdminURL("settings"), func(request Request) {
		user := request.GetUser()
		form := settingsForm(user).AddCSRFToken(request)

		request.SetData("admin_navigation_settings_selected", true)
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   app.getSettingsNavigation(user, "settings"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

	app.AdminController.Post(app.GetAdminURL("settings"), func(request Request) {
		validateCSRF(request)
		user := request.GetUser()
		form := settingsForm(user).AddCSRFToken(request)
		if form.Validate() {
			userResource, err := app.getResourceByItem(User{})
			if err != nil {
				panic(err)
			}
			must(userResource.bindData(&user, user, request.Params(), form.getFilter()))
			must(app.Save(&user))
			request.AddFlashMessage(messages.Get(getLocale(request), "admin_settings_changed"))
			request.Redirect(app.GetAdminURL("settings"))
			return
		}

		request.SetData("admin_navigation_settings_selected", true)
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   app.getSettingsNavigation(user, "settings"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

	changePasswordForm := func(request Request) *form {
		request.SetData("admin_navigation_settings_selected", true)
		user := request.GetUser()
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

	renderPasswordForm := func(request Request, form *form) {
		user := request.GetUser()
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   app.getSettingsNavigation(user, "password"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	app.AdminController.Get(app.GetAdminURL("password"), func(request Request) {
		form := changePasswordForm(request)
		renderPasswordForm(request, form)
	})

	app.AdminController.Post(app.GetAdminURL("password"), func(request Request) {
		form := changePasswordForm(request)
		form.BindData(request.Params())
		form.Validate()
		if form.Valid {
			password := request.Params().Get("newpassword")
			user := request.GetUser()
			must(user.newPassword(password))
			must(app.Save(&user))
			request.AddFlashMessage(messages.Get(getLocale(request), "admin_password_changed"))
			request.Redirect(app.GetAdminURL("settings"))
		} else {
			renderPasswordForm(request, form)
		}
	})

}
