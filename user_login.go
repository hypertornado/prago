package prago

import (
	"time"
)

func initUserLogin(resource *Resource) {

	resource.ItemAction("loginas").Name(unlocalized("Přihlásit se jako")).Permission(sysadminPermission).Handler(
		func(request *Request) {
			var user user
			resource.app.Is("id", request.Params().Get("id")).MustGet(&user)
			request.logInUser(&user)
			request.Redirect(resource.app.getAdminURL(""))
		},
	)

	resource.app.nologinFormAction("login", func(form *Form, request *Request) {
		locale := localeFromRequest(request)
		emailValue := request.Params().Get("email")
		emailInput := form.AddEmailInput("email", messages.Get(locale, "admin_email"))
		if emailValue == "" {
			emailInput.Focused = true
		}
		emailInput.Value = request.Params().Get("email")
		passwordInput := form.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		if emailValue != "" {
			passwordInput.Focused = true
		}
		form.AddSubmit(messages.Get(locale, "admin_login_action"))
	}, func(vc ValidationContext) {
		locale := vc.Locale()
		email := vc.GetValue("email")
		email = fixEmail(email)
		request := vc.Request()
		password := vc.GetValue("password")

		var user user
		err := request.app.Is("email", email).Get(&user)
		if err != nil {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		if !user.isPassword(password) {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		user.LoggedInTime = time.Now()
		user.LoggedInUseragent = request.Request().UserAgent()
		user.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

		must(request.app.Save(&user))
		request.logInUser(&user)
		request.AddFlashMessage(messages.Get(user.Locale, "admin_login_ok"))

		vc.Validation().RedirectionLocaliton = request.app.getAdminURL("")
	})

}
