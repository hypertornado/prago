package admin

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"time"
)

type User struct {
	ID                int64
	Name              string
	Email             string `prago-admin-unique:"true"`
	Password          string
	Locale            string
	IsSysadmin        bool
	IsAdmin           bool
	IsActive          bool
	LoggedInIP        string
	LoggedInUseragent string
	LoggedInTime      time.Time
	EmailConfirmedAt  time.Time
	EmailRenewedAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (User) AdminTableName() string { return "admin_user" }

func LoginForm(locale string) *Form {
	form := NewForm()
	form.Method = "POST"
	form.SubmitValue = messages.Messages.Get(locale, "admin_login_action")

	form.AddItem(&FormItem{
		Name:        "email",
		SubTemplate: "admin_item_email",
		NameHuman:   messages.Messages.Get(locale, "admin_email"),
	})

	form.AddItem(&FormItem{
		Name:        "email",
		SubTemplate: "admin_item_password",
		NameHuman:   messages.Messages.Get(locale, "admin_password"),
	})
	return form
}

func (User) AdminInitResource(a *Admin, resource *AdminResource) error {

	a.AdminAccessController.AddBeforeAction(func(request prago.Request) {
		request.SetData("locale", defaultLocale)
	})

	a.AdminAccessController.Get(a.GetURL(resource, "login"), func(request prago.Request) {
		locale := defaultLocale

		title := fmt.Sprintf("%s - %s", a.AppName, messages.Messages.Get(locale, "admin_login_name"))

		request.SetData("bottom", fmt.Sprintf("<a href=\"new\">%s</a><br><a href=\"forgot\">%s</a>",
			messages.Messages.Get(locale, "admin_register"),
			messages.Messages.Get(locale, "admin_forgoten"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", LoginForm(locale))
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	})

	a.AdminAccessController.Post(a.GetURL(resource, "login"), func(request prago.Request) {
		email := request.Params().Get("email")
		password := request.Params().Get("password")

		var user User
		err := a.Query().WhereIs("email", email).Get(&user)
		if err != nil {
			if err == ErrorNotFound {
				prago.Redirect(request, a.GetURL(resource, "login"))
			} else {
				panic(err)
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			prago.Redirect(request, a.GetURL(resource, "login"))
			return
		}

		session := request.GetData("session").(*sessions.Session)
		session.Values["user_id"] = user.ID

		prago.Must(session.Save(request.Request(), request.Response()))
		prago.Redirect(request, a.Prefix)
	})

	newUserForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
		form.SubmitValue = messages.Messages.Get(locale, "admin_register")
		form.AddTextInput("name", messages.Messages.Get(locale, "Name"),
			NonEmptyValidator(messages.Messages.Get(locale, "admin_user_name_not_empty")),
		)
		form.AddEmailInput("email", messages.Messages.Get(locale, "admin_email"),
			EmailValidator(messages.Messages.Get(locale, "admin_email_not_valid")),
			NewValidator(func(field *FormItem) bool {
				if len(field.Errors) != 0 {
					return true
				}
				var user User
				a.Query().WhereIs("email", field.Value).Get(&user)
				if user.Email == field.Value {
					return false
				}
				return true
			}, messages.Messages.Get(locale, "admin_email_already_registered")),
		)
		form.AddPasswordInput("password", messages.Messages.Get(locale, "admin_register_password"),
			MinLengthValidator("", 8),
		)
		return form
	}

	renderRegistration := func(request prago.Request, form *Form, locale string) {
		title := fmt.Sprintf("%s - %s", a.AppName, messages.Messages.Get(locale, "admin_register"))
		request.SetData("bottom", fmt.Sprintf("<a href=\"login\">%s</a>",
			messages.Messages.Get(locale, "admin_login_action"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "new"), func(request prago.Request) {
		locale := defaultLocale
		renderRegistration(request, newUserForm(locale), locale)
	})

	a.AdminAccessController.Post(a.GetURL(resource, "new"), func(request prago.Request) {
		locale := defaultLocale

		form := newUserForm(locale)

		form.BindData(request.Params())
		form.Validate()

		if form.Valid {
			email := request.Params().Get("email")
			password := request.Params().Get("password")

			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
			prago.Must(err)

			user := &User{}
			user.Email = email
			user.Password = string(passwordHash)

			prago.Must(a.Create(user))

			prago.Redirect(request, a.Prefix)

		} else {
			form.ItemMap["password"].Value = ""
			renderRegistration(request, form, locale)
		}
	})

	a.AdminAccessController.Get(a.Prefix+"/logout", func(request prago.Request) {
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "user_id")
		err := session.Save(request.Request(), request.Response())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.GetURL(resource, "login"))
	})

	a.AdminAccessController.Get(a.Prefix+"/admin.css", func(request prago.Request) {
		request.Response().Header().Add("Content-type", "text/css")
		request.SetData("statusCode", 200)
		request.SetData("body", []byte(CSS))
	})

	BindList(a, resource)

	return nil
}
