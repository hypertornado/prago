package prago

import (
	"fmt"
	"sort"
)

func initUserSettings(app *App) {

	ActionForm(app, "settings", func(form *Form, request *Request) {
		user := request.getUser()
		form.Title = messages.Get(request.Locale(), "admin_settings")

		name := form.AddTextInput("name", "")
		name.Name = messages.Get(request.Locale(), "Name")
		name.Value = user.Name

		phoneInput := form.AddTextInput("phone", "Telefon")
		phoneInput.Value = user.Phone

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

		phone := request.Param("phone")
		if phone != "" {
			if !IsPhoneNumberValid(phone) {
				vc.AddItemError("phone", "Neplatné telefonní číslo")
			}

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
			user.Phone = phone
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
	}).Icon("glyphicons-basic-5-settings.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_settings")).Board(app.optionsBoard)

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
		}).Icon("glyphicons-basic-45-key.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_password_change")).Board(app.optionsBoard)

	ActionResourceForm[user](app, "create", func(form *Form, request *Request) {

		cu := request.getUser()

		var roles []string
		for role := range app.accessManager.roles {
			if app.canManageRole(cu.Role, role) {
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			form.Title = "Nelze založit uživatele"
			return
		}

		sort.Strings(roles)

		var roleSelect [][2]string
		//roleSelect = append(roleSelect, [2]string{"", ""})
		for _, role := range roles {
			roleSelect = append(roleSelect, [2]string{role, role})
		}

		form.AddTextInput("username", "Uživatelské jméno")
		form.AddTextInput("name", "Jméno")
		form.AddTextInput("email", "Email")
		form.AddTextInput("phone", "Telefon")

		form.AddSelect("locale", messages.Get(request.Locale(), "admin_locale"), availableLocales)

		form.AddSelect("role", "Role", roleSelect)
		form.AddPasswordInput("password", "Heslo")
		form.AddSubmit("Založit uživatele")

	}, func(fv FormValidation, request *Request) {

		password := request.Param("password")
		role := request.Param("role")

		if role == "" {
			fv.AddItemError("role", "Vyberte roli")
		}

		cu := request.getUser()
		if !app.canManageRole(cu.Role, role) {
			fv.AddItemError("role", "Nemáte oprávnění pro zakládání této role")
		}

		if !isPasswordValid(password) {
			fv.AddItemError("password", "Krátké heslo, alespoň 7 znaků")
		}

		usr := &user{
			Username: request.Param("username"),
			Name:     request.Param("name"),
			Email:    request.Param("email"),
			Phone:    request.Param("phone"),
			Locale:   request.Param("locale"),
			Role:     role,
		}
		usr.newPassword(password)

		res, ok := TestValidationUpdate(app, usr, request)
		if !ok {
			for _, e := range res {
				fv.AddError(fmt.Sprintf("%s: %s", e.Field, e.Text))
			}
		}

		if !fv.Valid() {
			return
		}

		must(CreateWithLog(usr, request))
		request.AddFlashMessage("Uživatel založen")
		fv.Redirect(fmt.Sprintf("/admin/user/%d", usr.ID))

	}).Name(unlocalized("Založit uživatele")).Icon("glyphicons-basic-191-circle-empty-plus.svg")

	ActionResourceItemForm(app, "setup", func(user *user, form *Form, request *Request) {
		cu := request.getUser()
		if !app.canManageRole(cu.Role, user.Role) {
			form.Title = "Nemůžete upravovat tohoto uživatele"
			return
		}

		var roles []string
		roles = append(roles, "")
		for role := range app.accessManager.roles {
			roles = append(roles, role)
		}
		sort.Strings(roles)
		var roleSelect [][2]string
		for _, role := range roles {
			roleSelect = append(roleSelect, [2]string{role, role})
		}

		form.AddTextInput("username", "Uživatelské jméno").Value = user.Username
		form.AddTextInput("name", "Jméno").Value = user.Name
		form.AddTextInput("email", "Email").Value = user.Email
		form.AddTextInput("phone", "Telefon").Value = user.Phone

		form.AddSelect("role", "Role", roleSelect).Value = user.Role

		newPassword := form.AddPasswordInput("newpassword", messages.Get(request.Locale(), "admin_password_new"))
		newPassword.Focused = true
		newPassword.Autocomplete = "new-password"

		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(usr *user, fv FormValidation, request *Request) {

		cu := request.getUser()
		if !app.canManageRole(cu.Role, usr.Role) {
			panic("Nemůžete upravovat tohoto uživatele")
		}

		usr.Username = request.Param("username")
		usr.Name = request.Param("name")
		usr.Email = request.Param("email")
		usr.Phone = request.Param("phone")

		newRole := request.Param("role")
		if usr.Role != "" && !app.canManageRole(cu.Role, usr.Role) {
			fv.AddItemError("role", "Nemáte oprávnění měnit uživateli roli")
		}
		if newRole != "" && !app.canManageRole(cu.Role, newRole) {
			fv.AddItemError("role", "Nemáte oprávnění nastavit uživateli tuto roli")
		}
		usr.Role = newRole

		newpassword := request.Param("newpassword")
		if newpassword != "" {
			if isPasswordValid(newpassword) {
				must(usr.newPassword(newpassword))
			} else {
				fv.AddItemError("newpassword", messages.Get(request.Locale(), "admin_password_length"))
			}
		}

		res, ok := TestValidationUpdate(app, usr, request)
		if !ok {
			for _, e := range res {
				fv.AddError(fmt.Sprintf("%s: %s", e.Field, e.Text))
			}
		}

		if fv.Valid() {
			must(UpdateWithLog(usr, request))
			request.app.userDataCacheDelete(usr.ID)
			request.AddFlashMessage("Upraveno")
			fv.Redirect(fmt.Sprintf("/admin/user/%d", usr.ID))
		}

	}).Name(unlocalized("Nastavit uživatele"))

	ActionPlain(app, "logout", func(request *Request) {
		validateCSRF(request)
		request.logOutUser()
		request.Redirect(app.getAdminURL("login"))
	}).Icon("glyphicons-basic-432-log-out.svg").Permission(loggedPermission).Name(messages.GetNameFunction("admin_log_out")).Board(app.optionsBoard)
}

func isPasswordValid(in string) bool {
	if len(in) < 7 {
		return false
	}
	return true
}
