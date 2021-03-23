package prago

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//User represents admin user account
//TODO: better handle isactive user
type User struct {
	ID                int64
	Name              string `prago-preview:"true"`
	Email             string `prago-unique:"true" prago-preview:"true" prago-order:"true"`
	Role              string `prago-preview:"true" prago-type:"role"`
	Password          string `prago-can-view:"nobody"`
	Locale            string
	IsActive          bool
	LoggedInIP        string    `prago-can-view:"sysadmin"`
	LoggedInUseragent string    `prago-can-view:"sysadmin"`
	LoggedInTime      time.Time `prago-can-view:"sysadmin"`
	EmailConfirmedAt  time.Time `prago-can-view:"sysadmin"`
	EmailRenewedAt    time.Time `prago-can-view:"sysadmin"`
	CreatedAt         time.Time
	UpdatedAt         time.Time `prago-can-view:"sysadmin"`
}

//GetUser returns currently logged in user, it panics when there is no user
func (request Request) getUserOLD() User {
	u := request.GetData("currentuser").(*User)
	if u == nil {
		panic("no user found")
	}
	return *u
}

//TODO: remove
func basicUserAuthorize(request *Request) {
	if request.user.Role == "" {
		panic("can't authorize, user has no role")
	}
}

func fixEmail(in string) string {
	return strings.ToLower(in)
}

func (user User) gravatarURL() string {
	h := md5.New()
	io.WriteString(h, user.Email)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%ss.jpg?s=50&d=mp",
		fmt.Sprintf("%x", h.Sum(nil)),
	)
}

/*
func (user User) getRole() string {
	return user.Role
}*/

func (user *User) isPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	}
	return true
}

//TODO: better comparison
func (user *User) emailConfirmed() bool {
	if user.EmailConfirmedAt.Before(time.Now().AddDate(-1000, 0, 0)) {
		return false
	}
	return true
}

func (user *User) newPassword(password string) error {
	if len(password) < 7 {
		return errors.New("short password")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user.Password = string(passwordHash)
	return nil
}

func (user User) emailToken(app *App) string {
	randomness := app.ConfigurationGetString("random")
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", user.Email, randomness))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (app *App) initUserResource() {
	resource := app.Resource(User{})
	app.UsersResource = resource
	resource.name = messages.GetNameFunction("admin_users")
	//resource.PermissionView()
	resource.canEdit = sysadminPermission
	resource.canCreate = nobodyPermission
	resource.canDelete = sysadminPermission
	resource.canExport = sysadminPermission

	initUserRegistration(resource)
	initUserLogin(resource)
	resource.app.initUserSettings()
	initUserRenew(resource)
}
