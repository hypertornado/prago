package prago

import (
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/messages"
)

func initUserLogin(resource *Resource) {

	resource.AddItemAction(
		Action{
			Name:   func(string) string { return "Přihlásit se jako" },
			URL:    "loginas",
			Method: "get",
			Handler: func(resource Resource, request Request, u User) {
				if !u.IsSysadmin {
					panic("access denied")
				}

				id, err := strconv.Atoi(request.Params().Get("id"))
				if err != nil {
					panic(err)
				}

				var user User
				must(resource.App.Query().WhereIs("id", id).Get(&user))

				session := request.GetData("session").(*sessions.Session)
				session.Values["user_id"] = user.ID
				must(session.Save(request.Request(), request.Response()))
				request.Redirect(resource.App.GetURL(""))
			},
		})

	loginForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Messages.Get(locale, "admin_email")).Focused = true
		form.AddPasswordInput("password", messages.Messages.Get(locale, "admin_password"))
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_login_action"))
		return form
	}

	renderLogin := func(request Request, form *Form, locale string) {
		renderNavigationPageNoLogin(request, adminNavigationPage{
			Admin:        resource.App,
			Navigation:   resource.App.getNologinNavigation(locale, "login"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	resource.App.accessController.Get(resource.GetURL("login"), func(request Request) {
		locale := getLocale(request)
		form := loginForm(locale)
		renderLogin(request, form, locale)
	})

	resource.App.accessController.Post(resource.GetURL("login"), func(request Request) {
		email := request.Params().Get("email")
		email = fixEmail(email)
		password := request.Params().Get("password")

		session := request.GetData("session").(*sessions.Session)

		locale := getLocale(request)
		form := loginForm(locale)
		form.Items[0].Value = email
		form.Errors = []string{messages.Messages.Get(locale, "admin_login_error")}

		var user User
		err := resource.App.Query().WhereIs("email", email).Get(&user)
		if err != nil {
			if err == ErrItemNotFound {
				must(session.Save(request.Request(), request.Response()))
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

		must(resource.App.Save(&user))

		session.Values["user_id"] = user.ID
		session.AddFlash(messages.Messages.Get(locale, "admin_login_ok"))
		must(session.Save(request.Request(), request.Response()))
		request.Redirect(resource.App.GetURL(""))
	})

	resource.App.AdminController.Get(resource.App.GetURL("logout"), func(request Request) {
		ValidateCSRF(request)
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "user_id")
		session.AddFlash(messages.Messages.Get(getLocale(request), "admin_logout_ok"))
		must(session.Save(request.Request(), request.Response()))
		request.Redirect(resource.GetURL("login"))
	})

}
