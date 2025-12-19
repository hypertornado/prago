package prago

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//ALTER TABLE user DROP INDEX email;

var usernameRegex = regexp.MustCompile("^[a-z0-9.]{1,20}$")

type user struct {
	ID                int64 `prago-order-desc:"true"`
	Username          string
	Name              string
	Email             string
	Phone             string
	Role              string    `prago-type:"role"`
	Password          string    `prago-can-view:"nobody"`
	Locale            string    `prago-can-view:"sysadmin"`
	LoggedInIP        string    `prago-can-view:"sysadmin"`
	LoggedInUseragent string    `prago-can-view:"sysadmin"`
	LoggedInTime      time.Time `prago-can-view:"sysadmin"`
	EmailConfirmedAt  time.Time `prago-can-view:"sysadmin"`
	EmailRenewedAt    time.Time `prago-can-view:"sysadmin"`
	CreatedAt         time.Time
	UpdatedAt         time.Time `prago-can-view:"sysadmin"`
}

func fixEmail(in string) string {
	return strings.ToLower(in)
}

func (user *user) LongName() string {
	ret := user.Name
	if ret == "" {
		ret = user.Email
	}
	return ret
}

func (user *user) UserID() int64 {
	return user.ID
}

func (user *user) GetName(string) string {
	var ret string
	if user.Username != "" {
		ret = user.Username + " "
	}

	ret += fmt.Sprintf("%s %s", user.Name, user.Email)
	return ret
}

func (user *user) isPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	} else {
		return true
	}
}

// TODO: better comparison
func (user *user) emailConfirmed() bool {
	if user.EmailConfirmedAt.Before(time.Now().AddDate(-1000, 0, 0)) {
		return false
	} else {
		return true
	}
}

func (user *user) newPassword(password string) error {
	if !isPasswordValid(password) {
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
	randomness := app.mustGetSetting("random")
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", user.Email, randomness))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (app *App) initUserResource() {
	resource := NewResource[user](app)
	app.UsersResource = resource

	resource.Name(
		messages.GetNameFunction("admin_user"),
		messages.GetNameFunction("admin_users"),
	)
	resource.PermissionUpdate(sysadminPermission)
	resource.PermissionCreate(nobodyPermission)
	resource.PermissionDelete(sysadminPermission)
	resource.PermissionExport(sysadminPermission)

	resource.Icon("glyphicons-basic-4-user.svg")
	resource.Board(app.optionsBoard)
}

func (app *App) afterInitUserResource() {

	ValidateUpdate(app, func(usr *user, vc Validation, userData UserData) {

		if usr.Phone != "" {
			if !IsPhoneNumberValid(usr.Phone) {
				vc.AddItemError("phone", "Neplatný formát telefonního čísla")
			}
		}

		if usr.Email != "" {
			if !IsEmailValid(usr.Email) {
				vc.AddItemError("email", "Neplatný formát emailu")
			} else {
				sameEmailUsers := Query[user](app).Is("email", usr.Email).List()
				for _, same := range sameEmailUsers {
					if same.ID != usr.ID {
						vc.AddItemError("email", fmt.Sprintf("Duplicitní email s uživatelem #%d", same.ID))
					}
				}
			}
		}

		if usr.Username == "" && usr.Email == "" {
			vc.AddError("Musíte nastavit uživatelské jméno nebo email")
		}

		username := usr.Username
		if username != "" {
			if !usernameRegex.MatchString(username) {
				vc.AddItemError("username", "Špatný formát uživatelského jména")
			}

			var isUsed bool
			sameUsernameUsers := Query[user](app).Is("username", username).List()
			for _, sameUser := range sameUsernameUsers {
				if usr.ID != sameUser.ID {
					isUsed = true
				}
			}
			if isUsed {
				vc.AddItemError("username", fmt.Sprintf("Uživatelské jméno %s je již použito", username))
			}
		}

		if app.accessManager.roles[usr.Role] == nil {
			vc.AddItemError("role", "Neznámá role")
		}
	})

	initUserRegistration(app)
	initUserLogin(app)
	initUserSettings(app)
	initUserRenew(app)
}

func (app *App) GetCachedUserEmail(id int64) string {
	return <-Cached(app, fmt.Sprintf("cached-user-email-%d", id), func() string {
		user := Query[user](app).ID(id)
		if user == nil {
			return ""
		}
		return user.Email
	})
}
