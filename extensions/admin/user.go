package admin

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/sendgrid/sendgrid-go"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type User struct {
	ID                int64  `prago-preview:"false"`
	Name              string `prago-preview:"false"`
	Email             string `prago-unique:"true" prago-preview:"true" prago-order:"true"`
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

func (User) AdminName(lang string) string { return messages.Messages.Get(lang, "admin_users") }

func (u *User) AdminItemName(lang string) string {
	return u.Email
}

func (User) Authenticate(u *User) bool {
	return AuthenticateSysadmin(u)
}

func (u *User) IsPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func (u *User) EmailConfirmed() bool {
	if u.EmailConfirmedAt.Before(time.Now().AddDate(-1000, 0, 0)) {
		return false
	}
	return true
}

func (u *User) NewPassword(password string) error {
	if len(password) < 8 {
		return errors.New("short password")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	u.Password = string(passwordHash)
	return nil
}

func (u User) EmailToken(app *prago.App) string {
	randomness := app.Config().GetString("random")
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", u.Email, randomness))
	return fmt.Sprintf("%x", h.Sum(nil))
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
	formItem := form.AddHidden("_csrfToken")
	formItem.Value = CSRFToken(request)
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

func (u User) SendConfirmEmail(request prago.Request, a *Admin) error {

	if u.EmailConfirmed() {
		return errors.New("email already confirmed")
	}

	if a.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}

	locale := GetLocale(request)

	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.EmailToken(a.App))

	subject := messages.Messages.Get(locale, "admin_confirm_email_subject", a.AppName)
	link := request.App().Config().GetString("baseUrl") + a.Prefix + "/user/confirm_email?" + urlValues.Encode()
	body := messages.Messages.Get(locale, "admin_confirm_email_body", link, link, a.AppName)

	message := sendgrid.NewMail()
	message.SetFrom(a.noReplyEmail)
	message.AddTo(u.Email)
	message.AddToName(u.Name)
	message.SetSubject(subject)
	message.SetHTML(body)
	return a.sendgridClient.Send(message)
}

func (u User) SendAdminEmail(request prago.Request, a *Admin) error {
	if a.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}
	var users []*User
	err := a.Query().WhereIs("isadmin", true).Get(&users)
	if err != nil {
		return err
	}
	for _, user := range users {
		message := sendgrid.NewMail()
		message.SetFrom(a.noReplyEmail)
		message.AddTo(user.Email)
		message.AddToName(user.Name)
		message.SetSubject("New registration on " + a.AppName)
		message.SetHTML(fmt.Sprintf("New user registered on %s: %s (%s)", a.AppName, u.Email, u.Name))
		err = a.sendgridClient.Send(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u User) getRenewUrl(request prago.Request, a *Admin) string {
	urlValues := make(url.Values)
	urlValues.Add("email", u.Email)
	urlValues.Add("token", u.EmailToken(a.App))
	return request.App().Config().GetString("baseUrl") + a.Prefix + "/user/renew_password?" + urlValues.Encode()
}

func (u User) SendRenew(request prago.Request, a *Admin) error {
	if a.noReplyEmail == "" {
		return errors.New("no reply email empty")
	}

	locale := GetLocale(request)

	subject := messages.Messages.Get(locale, "admin_forgotten_email_subject", a.AppName)
	link := u.getRenewUrl(request, a)
	body := messages.Messages.Get(locale, "admin_forgotten_email_body", link, link, a.AppName)

	message := sendgrid.NewMail()
	message.SetFrom(a.noReplyEmail)
	message.AddTo(u.Email)
	message.AddToName(u.Name)
	message.SetSubject(subject)
	message.SetHTML(body)
	return a.sendgridClient.Send(message)
}

func (User) InitResource(a *Admin, resource *Resource) error {

	a.AdminAccessController.AddBeforeAction(func(request prago.Request) {
		request.SetData("locale", GetLocale(request))
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
		title := fmt.Sprintf("%s - %s", messages.Messages.Get(locale, "admin_login_name"), a.AppName)
		request.SetData("bottom", fmt.Sprintf("<a href=\"registration\">%s</a><br><a href=\"forgot\">%s</a>",
			messages.Messages.Get(locale, "admin_register"),
			messages.Messages.Get(locale, "admin_forgoten"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "confirm_email"), func(request prago.Request) {
		email := request.Params().Get("email")
		token := request.Params().Get("token")

		var user User
		err := a.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if !user.EmailConfirmed() {
				if token == user.EmailToken(request.App()) {
					user.EmailConfirmedAt = time.Now()
					err = a.Save(&user)
					if err == nil {
						FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_confirm_email_ok"))
						prago.Redirect(request, a.Prefix+"/user/login")
						return
					}
				}
			}
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_confirm_email_fail"))
		prago.Redirect(request, a.Prefix+"/user/login")
	})

	forgotForm := func(locale string) *Form {
		form := NewForm()
		form.Method = "POST"
		form.AddEmailInput("email", messages.Messages.Get(locale, "admin_email")).Focused = true
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_forgotten_submit"))
		return form
	}

	renderForgot := func(request prago.Request, form *Form, locale string) {
		title := fmt.Sprintf("%s - %s", messages.Messages.Get(locale, "admin_forgotten_name"), a.AppName)
		request.SetData("bottom", fmt.Sprintf("<a href=\"login\">%s</a>",
			messages.Messages.Get(locale, "admin_login_action"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "forgot"), func(request prago.Request) {
		locale := GetLocale(request)
		form := forgotForm(locale)
		renderForgot(request, form, locale)
	})

	a.AdminAccessController.Post(a.GetURL(resource, "forgot"), func(request prago.Request) {
		email := request.Params().Get("email")
		email = FixEmail(email)

		var user User
		err := a.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if user.EmailConfirmed() {
				if !time.Now().AddDate(0, 0, -1).Before(user.EmailRenewedAt) {
					user.EmailRenewedAt = time.Now()
					err = a.Save(&user)
					if err == nil {
						err = user.SendRenew(request, a)
						if err == nil {
							FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_forgoten_sent", user.Email))
							prago.Redirect(request, a.Prefix+"/user/login")
							return
						}
					}
				}
			}
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_forgoten_error", user.Email))
		prago.Redirect(request, a.Prefix+"/user/forgot")
	})

	renewPasswordForm := func(locale string) (form *Form) {
		form = NewForm()
		form.Method = "POST"

		passwordInput := form.AddPasswordInput("password", messages.Messages.Get(locale, "admin_password_new"),
			MinLengthValidator(messages.Messages.Get(locale, "admin_password_length"), 8))
		passwordInput.Focused = true
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_forgoten_set"))
		return
	}

	renderRenew := func(request prago.Request, form *Form, locale string) {
		email := request.Params().Get("email")
		email = FixEmail(email)
		title := fmt.Sprintf("%s - %s", email, messages.Messages.Get(locale, "admin_forgoten_set"))
		request.SetData("bottom", fmt.Sprintf("<a href=\"login\">%s</a>",
			messages.Messages.Get(locale, "admin_login_action"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "renew_password"), func(request prago.Request) {
		locale := GetLocale(request)
		form := renewPasswordForm(locale)
		renderRenew(request, form, locale)
	})

	a.AdminAccessController.Post(a.GetURL(resource, "renew_password"), func(request prago.Request) {
		locale := GetLocale(request)

		form := renewPasswordForm(locale)

		form.BindData(request.Params())
		form.Validate()

		email := request.Params().Get("email")
		email = FixEmail(email)
		token := request.Params().Get("token")

		errStr := messages.Messages.Get(locale, "admin_error")

		var user User
		err := a.Query().WhereIs("email", email).Get(&user)
		if err == nil {
			if token == user.EmailToken(request.App()) {
				if form.Valid {
					err = user.NewPassword(request.Params().Get("password"))
					if err == nil {
						err = a.Save(&user)
						if err == nil {
							FlashMessage(request, messages.Messages.Get(locale, "admin_password_changed"))
							prago.Redirect(request, a.Prefix+"/user/login")
							return
						}
					}
				}
			}
		}
		FlashMessage(request, errStr)
		form.GetItemByName("password").Value = ""
		renderLogin(request, form, locale)
	})

	a.AdminAccessController.Get(a.GetURL(resource, "login"), func(request prago.Request) {
		locale := GetLocale(request)
		form := loginForm(locale)
		renderLogin(request, form, locale)
	})

	a.AdminAccessController.Post(a.GetURL(resource, "login"), func(request prago.Request) {
		email := request.Params().Get("email")
		email = FixEmail(email)
		password := request.Params().Get("password")

		session := request.GetData("session").(*sessions.Session)

		locale := GetLocale(request)
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
			}
			panic(err)
		}

		if !user.IsPassword(password) {
			renderLogin(request, form, locale)
			return
		}

		user.LoggedInTime = time.Now()
		user.LoggedInUseragent = request.Request().UserAgent()
		user.LoggedInIP = request.Request().Header.Get("X-Forwarded-For")

		prago.Must(a.Save(&user))

		session.Values["user_id"] = user.ID
		session.AddFlash(messages.Messages.Get(locale, "admin_login_ok"))
		prago.Must(session.Save(request.Request(), request.Response()))
		prago.Redirect(request, a.Prefix)
	})

	a.AdminController.Get(a.GetURL(resource, "as")+"/:id", func(request prago.Request) {

		u := GetUser(request)
		if !u.IsSysadmin {
			panic("access denied")
		}

		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		var user User
		prago.Must(a.Query().WhereIs("id", id).Get(&user))

		session := request.GetData("session").(*sessions.Session)
		session.Values["user_id"] = user.ID
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
		form.AddSubmit("send", messages.Messages.Get(locale, "admin_register"))
		return form
	}

	renderRegistration := func(request prago.Request, form *Form, locale string) {
		title := fmt.Sprintf("%s - %s", messages.Messages.Get(locale, "admin_register"), a.AppName)
		request.SetData("bottom", fmt.Sprintf("<a href=\"login\">%s</a>",
			messages.Messages.Get(locale, "admin_login_action"),
		))
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("admin_form", form)
		request.SetData("title", title)

		prago.Render(request, 200, "admin_login")
	}

	a.AdminAccessController.Get(a.GetURL(resource, "registration"), func(request prago.Request) {
		locale := GetLocale(request)
		renderRegistration(request, newUserForm(locale), locale)
	})

	a.AdminAccessController.Post(a.GetURL(resource, "registration"), func(request prago.Request) {
		locale := GetLocale(request)

		form := newUserForm(locale)

		form.BindData(request.Params())
		form.Validate()

		if form.Valid {
			email := request.Params().Get("email")
			email = FixEmail(email)
			user := &User{}
			user.Email = email
			user.Name = request.Params().Get("name")
			user.IsActive = true
			user.Locale = locale
			prago.Must(user.NewPassword(request.Params().Get("password")))
			prago.Must(user.SendConfirmEmail(request, a))
			err := user.SendAdminEmail(request, a)
			if err != nil {
				request.App().Log().Println(err)
			}
			prago.Must(a.Create(user))

			FlashMessage(request, messages.Messages.Get(locale, "admin_confirm_email_send", user.Email))
			prago.Redirect(request, a.Prefix+"/user/login")
		} else {
			form.GetItemByName("password").Value = ""
			renderRegistration(request, form, locale)
		}
	})

	a.AdminController.Get(a.Prefix+"/logout", func(request prago.Request) {
		ValidateCSRF(request)
		session := request.GetData("session").(*sessions.Session)
		delete(session.Values, "user_id")
		session.AddFlash(messages.Messages.Get(GetLocale(request), "admin_logout_ok"))
		err := session.Save(request.Request(), request.Response())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.GetURL(resource, "login"))
	})

	settingsForm := func(locale string, user *User) *Form {
		form, err := resource.StructCache.GetForm(user, locale, WhiteListFilter("Name", "Email"), WhiteListFilter("Name", "Locale"))
		if err != nil {
			panic(err)
		}

		sel := form.AddSelect("Locale", messages.Messages.Get(locale, "admin_locale"), availableLocales)
		sel.Value = user.Locale

		form.AddSubmit("_submit", messages.Messages.Get(locale, "admin_edit"))
		return form
	}

	renderSettings := func(request prago.Request, user *User, form *Form) {
		request.SetData("admin_item", user)
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_settings")
		prago.Render(request, 200, "admin_layout")
	}

	a.AdminController.Get(a.GetURL(resource, "settings"), func(request prago.Request) {
		user := GetUser(request)
		form := settingsForm(GetLocale(request), user)
		AddCSRFToken(form, request)
		request.SetData("admin_title", messages.Messages.Get(GetLocale(request), "admin_settings"))
		renderSettings(request, user, form)
	})

	a.AdminController.Post(a.GetURL(resource, "settings"), func(request prago.Request) {
		ValidateCSRF(request)
		user := GetUser(request)
		form := settingsForm(GetLocale(request), user)
		AddCSRFToken(form, request)
		form.Validate()
		if form.Valid {
			prago.Must(resource.StructCache.BindData(user, request.Params(), request.Request().MultipartForm, form.GetFilter()))
			prago.Must(resource.Save(user))
			FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_settings_changed"))
			prago.Redirect(request, a.GetURL(resource, "settings"))
			return
		}
		renderSettings(request, user, form)
	})

	changePasswordForm := func(request prago.Request) *Form {
		user := GetUser(request)
		locale := GetLocale(request)
		oldValidator := NewValidator(func(field *FormItem) bool {
			if !user.IsPassword(field.Value) {
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
			MinLengthValidator(messages.Messages.Get(locale, "admin_password_length"), 8),
		)
		form.AddSubmit("_submit", messages.Messages.Get(locale, "admin_save"))
		return form
	}

	renderPasswordForm := func(request prago.Request, form *Form) {
		request.SetData("admin_title", messages.Messages.Get(GetLocale(request), "admin_password_change"))
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_form_view")
		prago.Render(request, 200, "admin_layout")
	}

	a.AdminController.Get(a.GetURL(resource, "password"), func(request prago.Request) {
		form := changePasswordForm(request)
		renderPasswordForm(request, form)
	})

	a.AdminController.Post(a.GetURL(resource, "password"), func(request prago.Request) {
		form := changePasswordForm(request)
		form.BindData(request.Params())
		form.Validate()
		if form.Valid {
			password := request.Params().Get("newpassword")
			user := GetUser(request)
			prago.Must(user.NewPassword(password))
			prago.Must(resource.Save(user))
			FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_password_changed"))
			prago.Redirect(request, a.GetURL(resource, "settings"))
		} else {
			renderPasswordForm(request, form)
		}
	})

	return nil
}

func FixEmail(in string) string {
	return strings.ToLower(in)
}
