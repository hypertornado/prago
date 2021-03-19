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
type User struct {
	ID       int64
	Name     string `prago-preview:"true"`
	Email    string `prago-unique:"true" prago-preview:"true" prago-order:"true"`
	Role     string `prago-preview:"true" prago-type:"role" prago-description:"Role"`
	Password string `prago-view:"nobody"`
	Locale   string
	//IsSysadmin        bool `prago-preview:"true" prago-description:"Sysadmin"`
	IsActive          bool
	LoggedInIP        string    `prago-view:"sysadmin"`
	LoggedInUseragent string    `prago-view:"sysadmin"`
	LoggedInTime      time.Time `prago-view:"sysadmin"`
	EmailConfirmedAt  time.Time `prago-view:"sysadmin"`
	EmailRenewedAt    time.Time `prago-view:"sysadmin"`
	CreatedAt         time.Time
	UpdatedAt         time.Time `prago-view:"sysadmin"`
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

func initUserResource(resource *Resource) {
	resource.name = messages.GetNameFunction("admin_users")
	resource.canEdit = sysadminPermission

	initUserRegistration(resource)
	initUserLogin(resource)
	resource.app.initUserSettings()
	initUserRenew(resource)
}
