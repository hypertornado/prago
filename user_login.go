package prago

import (
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

func initUserLogin(resource *Resource) {

	resource.ItemAction("loginas").Name(Unlocalized("Přihlásit se jako")).Permission(permissionSysadmin).Handler(
		func(request Request) {
			u := request.getUser()
			if !u.IsSysadmin {
				panic("access denied")
			}

			id, err := strconv.Atoi(request.Params().Get("id"))
			if err != nil {
				panic(err)
			}

			var user User
			must(resource.app.Query().WhereIs("id", id).Get(&user))

			session := request.GetData("session").(*sessions.Session)
			session.Values["user_id"] = user.ID
			must(session.Save(request.Request(), request.Response()))
			request.Redirect(resource.app.getAdminURL(""))
		},
	)

	loginForm := func(locale string) *form {
		form := newForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Get(locale, "admin_email")).Focused = true
		form.AddPasswordInput("password", messages.Get(locale, "admin_password"))
		form.AddSubmit("send", messages.Get(locale, "admin_login_action"))
		return form
	}

	renderLogin := func(request Request, form *form, locale string) {
		renderNavigationPageNoLogin(request, page{
			App:          resource.app,
			Navigation:   resource.app.getNologinNavigation(locale, "login"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	}

	resource.app.accessController.get(resource.getURL("login"), func(request Request) {
		locale := getLocale(request)
		form := loginForm(locale)
		renderLogin(request, form, locale)
	})

	resource.app.accessController.post(resource.getURL("login"), func(request Request) {
		email := request.Params().Get("email")
		email = fixEmail(email)
		password := request.Params().Get("password")

		session := request.GetData("session").(*sessions.Session)

		locale := getLocale(request)
		form := loginForm(locale)
		form.Items[0].Value = email
		form.Errors = []string{messages.Get(locale, "admin_login_error")}

		var user User
		err := resource.app.Query().WhereIs("email", email).Get(&user)
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

		must(resource.app.Save(&user))

		session.Values["user_id"] = user.ID
		session.AddFlash(messages.Get(locale, "admin_login_ok"))
		must(session.Save(request.Request(), request.Response()))
		request.Redirect(resource.app.getAdminURL(""))
	})
}
