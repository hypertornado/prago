package prago

import "github.com/hypertornado/prago/messages"

func initUserSettings(resource *Resource) {

	settingsForm := func(user User) *Form {
		form := NewForm()
		form.Method = "POST"
		form.Action = "settings"

		name := form.AddTextInput("name", "")
		name.NameHuman = messages.Messages.Get(user.Locale, "Name")
		name.Value = user.Name

		sel := form.AddSelect("locale", messages.Messages.Get(user.Locale, "admin_locale"), availableLocales)
		sel.Value = user.Locale

		form.AddSubmit("_submit", messages.Messages.Get(user.Locale, "admin_edit"))
		return form
	}

	resource.App.AdminController.Get(resource.GetURL("settings"), func(request Request) {
		user := request.GetUser()
		form := settingsForm(user).AddCSRFToken(request)

		request.SetData("admin_navigation_settings_selected", true)
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   resource.App.getSettingsNavigation(user, "settings"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

	resource.App.AdminController.Post(resource.GetURL("settings"), func(request Request) {
		ValidateCSRF(request)
		user := request.GetUser()
		form := settingsForm(user).AddCSRFToken(request)
		if form.Validate() {
			must(resource.bindData(&user, user, request.Params(), form.getFilter()))
			must(resource.App.Save(&user))
			AddFlashMessage(request, messages.Messages.Get(getLocale(request), "admin_settings_changed"))
			request.Redirect(resource.GetURL("settings"))
			return
		}

		request.SetData("admin_navigation_settings_selected", true)
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   resource.App.getSettingsNavigation(user, "settings"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	})

	changePasswordForm := func(request Request) *Form {
		request.SetData("admin_navigation_settings_selected", true)
		user := request.GetUser()
		locale := getLocale(request)
		oldValidator := NewValidator(func(field *FormItem) bool {
			if !user.isPassword(field.Value) {
				return false
			}
			return true
		}, messages.Messages.Get(locale, "admin_password_wrong"))

		form := NewForm()
		form.Method = "POST"
		form.AddPasswordInput("oldpassword",
			messages.Messages.Get(locale, "admin_password_old"),
			oldValidator,
		)
		form.AddPasswordInput("newpassword",
			messages.Messages.Get(locale, "admin_password_new"),
			MinLengthValidator(messages.Messages.Get(locale, "admin_password_length"), 7),
		)
		form.AddSubmit("_submit", messages.Messages.Get(locale, "admin_save"))
		return form
	}

	renderPasswordForm := func(request Request, form *Form) {
		user := request.GetUser()
		renderNavigationPage(request, adminNavigationPage{
			Navigation:   resource.App.getSettingsNavigation(user, "password"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	resource.App.AdminController.Get(resource.GetURL("password"), func(request Request) {
		form := changePasswordForm(request)
		renderPasswordForm(request, form)
	})

	resource.App.AdminController.Post(resource.GetURL("password"), func(request Request) {
		form := changePasswordForm(request)
		form.BindData(request.Params())
		form.Validate()
		if form.Valid {
			password := request.Params().Get("newpassword")
			user := request.GetUser()
			must(user.newPassword(password))
			must(resource.App.Save(&user))
			AddFlashMessage(request, messages.Messages.Get(getLocale(request), "admin_password_changed"))
			request.Redirect(resource.GetURL("settings"))
		} else {
			renderPasswordForm(request, form)
		}
	})

}
