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

		var u user
		err := request.app.Is("email", email).Get(&u)
		if err != nil {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		if !u.isPassword(password) {
			vc.AddError(messages.Get(locale, "admin_login_error"))
			return
		}

		u.LoggedInTime = time.Now()
		u.LoggedInUseragent = request.Request().UserAgent()
		u.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

		must(GetResource[user](resource.app).Update(&u))
		//must(request.app.Save(&user))
		request.logInUser(&u)
		request.AddFlashMessage(messages.Get(u.Locale, "admin_login_ok"))

		vc.Validation().RedirectionLocaliton = request.app.getAdminURL("")
	})

}
