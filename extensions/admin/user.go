package admin

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"time"
)

type User struct {
	ID                int64
	Name              string
	Email             string
	Password          string
	Locale            string
	IsActive          bool
	LoggedInTime      time.Time
	LoggedInIP        string
	LoggedInUseragent string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (User) AdminTableName() string { return "admin_user" }

func (User) AdminInitResource(a *Admin, resource *AdminResource) error {

	a.AdminAccessController.AddBeforeAction(func(request prago.Request) {
		request.SetData("locale", defaultLocale)
	})

	a.AdminAccessController.Get(a.GetURL(resource, "login"), func(request prago.Request) {
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("name", a.AppName)
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

	a.AdminAccessController.Get(a.GetURL(resource, "new"), func(request prago.Request) {
		request.SetData("admin_header_prefix", a.Prefix)
		request.SetData("name", a.AppName)
		prago.Render(request, 200, "admin_new_user")
	})

	a.AdminAccessController.Post(a.GetURL(resource, "new"), func(request prago.Request) {
		email := request.Params().Get("email")
		password := request.Params().Get("password")

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		prago.Must(err)

		user := &User{}
		user.Email = email
		user.Password = string(passwordHash)

		prago.Must(a.Create(user))

		prago.Redirect(request, a.Prefix)

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
