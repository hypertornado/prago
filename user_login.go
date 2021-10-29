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

	/*loginForm := func(request *Request) *formView {
		locale := localeFromRequest(request)
		form := newForm()
		formView := form.GetFormView(request)
		formView.AJAX = true
		formView.Classes = append(formView.Classes, "prago_form")
		formView.Method = "POST"
		formView.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		formView.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		formView.AddSubmit("send", messages.Get(locale, "admin_login_action"))
		return formView
	}

	renderLogin := func(request *Request, form *formView, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          resource.app,
			Navigation:   resource.app.getNologinNavigation(locale, "login"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}*/

	resource.app.accessController.get(resource.getURL("login"), func(request *Request) {
		if request.user != nil {
			request.Redirect("/admin")
			return
		}

		locale := localeFromRequest(request)
		form := newForm()
		form.AJAX = true
		form.Action = "/admin/user/login"
		formView := form.GetFormView(request)
		formView.Classes = append(formView.Classes, "prago_form")

		emailValue := request.Params().Get("email")

		emailInput := formView.AddEmailInput("email", messages.Get(locale, "admin_email"))
		if emailValue == "" {
			emailInput.Focused = true
		}
		emailInput.Value = request.Params().Get("email")
		passwordInput := formView.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		if emailValue != "" {
			passwordInput.Focused = true
		}

		formView.AddSubmit("send", messages.Get(locale, "admin_login_action"))

		renderNavigationPageNoLogin(request, page{
			App:          resource.app,
			Navigation:   resource.app.getNologinNavigation(locale, "login"),
			PageTemplate: "admin_form",
			PageData:     formView,
		})

		//locale := localeFromRequest(request)
		//form := loginForm(request)
		//renderLogin(request, form, locale)
	})

	resource.app.accessController.post(resource.getURL("login"), func(request *Request) {
		request.RenderJSON(
			loginValidation(request),
		)
	})

	/*

		resource.app.accessController.post(resource.getURL("login"), func(request *Request) {
			email := request.Params().Get("email")
			email = fixEmail(email)
			password := request.Params().Get("password")

			locale := localeFromRequest(request)
			form := loginForm(request)
			form.Items[0].Value = email
			form.Errors = []string{messages.Get(locale, "admin_login_error")}

			var user user
			err := resource.app.Is("email", email).Get(&user)
			if err != nil {
				if err == ErrItemNotFound {
					renderLogin(request, form, locale)
					return
				}
				panic(err)
			}

			if !user.isPassword(password) {
				renderLogin(request, form, locale)
				return
			}

			user.LoggedInTime = time.Now()
			user.LoggedInUseragent = request.Request().UserAgent()
			user.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

			must(resource.app.Save(&user))
			request.logInUser(&user)
			request.AddFlashMessage(messages.Get(locale, "admin_login_ok"))
			request.Redirect(resource.app.getAdminURL(""))
		})*/
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
