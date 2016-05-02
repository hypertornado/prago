package admin

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"io"
	"time"
)

type User struct {
	ID                int64  `prago-preview:"false"`
	Name              string `prago-preview:"false"`
	Email             string `prago-unique:"true" prago-preview:"true"`
	Password          string
	Locale            string
	IsSysadmin        bool `prago-preview:"true" prago-description:"Sysadmin"`
	IsAdmin           bool `prago-preview:"true" prago-description:"Admin"`
	IsActive          bool
	LoggedInIP        string
	LoggedInUseragent string
	LoggedInTime      time.Time
	EmailConfirmedAt  time.Time
	EmailRenewedAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (User) Authenticate(u User) bool {
	return AuthenticateSysadmin(u)
}

func (u *User) CSRFToken(randomness string) string {
	if len(randomness) <= 0 {
		panic("randomness too short")
	}

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%s%s", u.ID, randomness, u.LoggedInTime))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CSRFToken(request prago.Request) string {
	return request.GetData("_csrfToken").(string)
}

func AddCSRFToken(form *Form, request prago.Request) {
	form.AddHidden("_csrfToken", CSRFToken(request))
}

func ValidateCSRF(request prago.Request) {
	token := CSRFToken(request)
	if len(token) == 0 {
		panic("token not set")
	}
	paramsToken := request.Params().Get("_csrfToken")
	if paramsToken != token {
		panic("Wrong CSRF token")
	}
}

func (User) AdminTableName() string { return "admin_user" }

func (User) AdminInitResource(a *Admin, resource *AdminResource) error {

	a.AdminAccessController.AddBeforeAction(func(request prago.Request) {
		request.SetData("locale", defaultLocale)
	})

	loginForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Messages.Get(locale, "admin_email")).Focused = true
		form.AddPasswordInput("password", messages.Messages.Get(locale, "admin_password"))
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_login_action"))
		return form
	}

	renderLogin := func(request prago.Request, form *Form, locale string) {
		title := fmt.Sprintf("%s - %s", a.AppName, messages.Messages.Get(locale, "admin_login_name"))
		request.SetData("bottom", fmt.Sprintf("<a href=\"new\">%s</a><br><a href=\"forgot\">%s</a>",
			messages.Messages.Get(locale, "admin_register"),
			messages.Messages.Get(locale, "admin_forgoten"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "login"), func(request prago.Request) {
		locale := defaultLocale
		form := loginForm(locale)
		renderLogin(request, form, locale)

	})

	a.AdminAccessController.Post(a.GetURL(resource, "login"), func(request prago.Request) {
		email := request.Params().Get("email")
		password := request.Params().Get("password")

		session := request.GetData("session").(*sessions.Session)

		locale := defaultLocale
		form := loginForm(locale)
		form.Items[0].Value = email
		form.Errors = []string{messages.Messages.Get(locale, "admin_login_error")}

		var user User
		err := a.Query().WhereIs("email", email).Get(&user)
		if err != nil {
			if err == ErrorNotFound {
				prago.Must(session.Save(request.Request(), request.Response()))
				renderLogin(request, form, locale)
				return
			} else {
				panic(err)
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			renderLogin(request, form, locale)
			return
		}

		session.Values["user_id"] = user.ID
		session.AddFlash(messages.Messages.Get(locale, "admin_login_ok"))
		prago.Must(session.Save(request.Request(), request.Response()))
		prago.Redirect(request, a.Prefix)
	})

	newUserForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
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
		//form.AddCheckbox("ok", "XX")
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_register"))
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

	a.AdminAccessController.Get(a.Prefix+"/admin.css", func(request prago.Request) {
		request.Response().Header().Add("Content-type", "text/css")
		request.SetData("statusCode", 200)
		request.SetData("body", []byte(CSS))
	})

	a.AdminController.Get(a.Prefix+"/logout", func(request prago.Request) {
		ValidateCSRF(request)
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "user_id")
		session.AddFlash(messages.Messages.Get(defaultLocale, "admin_logout_ok"))
		err := session.Save(request.Request(), request.Response())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.GetURL(resource, "login"))
	})

	//BindList(a, resource)

	prago.Must(AdminInitResourceDefault(a, resource))

	return nil
}
