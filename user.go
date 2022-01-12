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
type user struct {
	ID                int64     `prago-order-desc:"true"`
	Name              string    `prago-preview:"true"`
	Email             string    `prago-unique:"true" prago-preview:"true"`
	Role              string    `prago-preview:"true" prago-type:"role"`
	Password          string    `prago-can-view:"nobody"`
	Locale            string    `prago-can-view:"sysadmin"`
	IsActive          bool      `prago-preview:"true"`
	LoggedInIP        string    `prago-can-view:"sysadmin" prago-preview:"true"`
	LoggedInUseragent string    `prago-can-view:"sysadmin" prago-preview:"true"`
	LoggedInTime      time.Time `prago-can-view:"sysadmin"`
	EmailConfirmedAt  time.Time `prago-can-view:"sysadmin"`
	EmailRenewedAt    time.Time `prago-can-view:"sysadmin"`
	CreatedAt         time.Time
	UpdatedAt         time.Time `prago-can-view:"sysadmin" prago-preview:"true"`
}

func fixEmail(in string) string {
	return strings.ToLower(in)
}

func (user *user) isPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	} else {
		return true
	}
}

//TODO: better comparison
func (user *user) emailConfirmed() bool {
	if user.EmailConfirmedAt.Before(time.Now().AddDate(-1000, 0, 0)) {
		return false
	} else {
		return true
	}
}

func (user *user) newPassword(password string) error {
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

func (user user) emailToken(app *App) string {
	randomness := app.ConfigurationGetString("random")
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", user.Email, randomness))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (app *App) initUserResource() {
	resource := NewResource[user](app).Resource
	app.UsersResource = resource

	resource.name = messages.GetNameFunction("admin_users")
	resource.canEdit = sysadminPermission
	resource.canCreate = nobodyPermission
	resource.canDelete = sysadminPermission
	resource.canExport = sysadminPermission

	initUserRegistration(resource)
	initUserLogin(GetResource[user](app))
	resource.app.initUserSettings()
	initUserRenew(resource)
}

func (app *App) GetCachedUserEmail(id int64) string {
	return app.Cache.Load(fmt.Sprintf("cached-user-email-%d", id), func() interface{} {
		var user user
		app.Is("id", id).Get(&user)
		return user.Email
	}).(string)
}
