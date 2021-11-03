package prago

import (
	"strconv"
	"time"
)

func initUserLogin(resource *Resource) {

	resource.ItemAction("loginas").Name(unlocalized("Přihlásit se jako")).Permission(sysadminPermission).Handler(
		func(request *Request) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var user user
			resource.app.Is("id", id).MustGet(&user)
			request.logInUser(&user)
			request.Redirect(resource.app.getAdminURL(""))
		},
	)

	resource.app.accessController.get(resource.getURL("login"), func(request *Request) {
		/*if request.user != nil {
			request.Redirect("/admin")
			return
		}*/

		locale := localeFromRequest(request)
		form := NewForm("/admin/user/login")

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

		form.AddSubmit("send", messages.Get(locale, "admin_login_action"))

		renderNavigationPageNoLogin(request, page{
			App:          resource.app,
			Navigation:   resource.app.getNologinNavigation(locale, "login"),
			PageTemplate: "admin_form",
			PageData:     form,
		})

	})

	resource.app.accessController.post(resource.getURL("login"), func(request *Request) {
		request.RenderJSON(
			loginValidation(request),
		)
	})

}

func loginValidation(request *Request) *FormValidation {
	locale := localeFromRequest(request)
	ret := NewFormValidation()
	email := request.Params().Get("email")
	email = fixEmail(email)
	password := request.Params().Get("password")

	var user user
	err := request.app.Is("email", email).Get(&user)
	if err != nil {
		ret.Errors = append(ret.Errors, FormValidationError{
			Text: messages.Get(locale, "admin_login_error"),
		})
		return ret
	}

	if !user.isPassword(password) {
		ret.Errors = append(ret.Errors, FormValidationError{
			Text: messages.Get(locale, "admin_login_error"),
		})
		return ret
	}

	user.LoggedInTime = time.Now()
	user.LoggedInUseragent = request.Request().UserAgent()
	user.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

	must(request.app.Save(&user))
	request.logInUser(&user)
	request.AddFlashMessage(messages.Get(locale, "admin_login_ok"))

	ret.RedirectionLocaliton = request.app.getAdminURL("")
	return ret

}
