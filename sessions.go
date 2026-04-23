package prago

import (
	"fmt"
	"time"
)

type session struct {
	ID         int64 `prago-order-desc:"true"`
	UUID       string
	User       int64     `prago-type:"relation"`
	UserAgent  string    `prago-type:"text" prago-name:"User Agent"`
	IPAddress  string    `prago-type:"text" prago-name:"IP Adresa"`
	LastAccess time.Time `prago-type:"timestamp"  prago-name:"Poslední přístup"`
	IsAPI      bool
	IsDeleted  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (app *App) initSessionsResource() {
	app.sessionsResource = NewResource[session](app).PermissionView("sysadmin").Name(unlocalized("Session"), unlocalized("Sessiony")).Board(sysadminBoard)
	app.initSessionsCache()

	ActionResourceItemForm(app, "logout", func(ses *session, form *Form, request *Request) {
		form.AddSubmit("Odhlásit")
	}, func(ses *session, fv FormValidation, request *Request) {

		err := app.deleteSession(ses.UUID)
		if err != nil {
			fv.AddError(err.Error())
		} else {
			fv.AddOK("Session odhlášena")
		}
	}).Name(unlocalized("Odhlásit"))
}

func generateSessionKey(isAPI bool) string {

	prefix := "PSK_"
	if isAPI {
		prefix = "PAK_"
	}
	return fmt.Sprintf("%s%s", prefix, randomString(64))
}

func (app *App) createSessionKey(user *user, isAPI bool) string {
	ses := &session{
		UUID:  generateSessionKey(isAPI),
		User:  user.ID,
		IsAPI: isAPI,
	}
	must(CreateItem(app, ses))
	return ses.UUID
}

func (app *App) getUserIDFromSession(sessionID string, api bool) int64 {
	id := app.getSessionCacheUserID(sessionID)
	if id > 0 {
		return id
	}

	ses := Query[session](app).Is("isdeleted", false).Is("uuid", sessionID).Is("isapi", api).First()
	if ses == nil {
		return 0
	}
	app.setSessionCacheUserID(ses.UUID, ses.User)

	return ses.User
}

func (app *App) deleteSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("Nelze smazat prazdnou session")
	}

	ses := Query[session](app).Is("uuid", sessionID).First()
	if ses == nil {
		return fmt.Errorf("Nelze najít session")
	}

	ses.IsDeleted = true
	err := UpdateItem(app, ses)
	if err != nil {
		return err
	}

	app.deleteSessionCacheUserID(sessionID)

	return nil
}
