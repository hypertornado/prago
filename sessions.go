package prago

import (
	"fmt"
	"time"
)

type session struct {
	ID         int64 `prago-order-desc:"true"`
	UUID       string
	User       int64  `prago-type:"relation"`
	UserAgent  string `prago-type:"text"`
	IPAddress  string `prago-type:"text"`
	LastAccess time.Time
	IsDeleted  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (app *App) initSessionsResource() {
	NewResource[session](app).PermissionView("sysadmin").Name(unlocalized("Session"), unlocalized("Sessiony")).Board(sysadminBoard)
	app.initSessionsCache()
}

func generateSessionKey() string {
	return fmt.Sprintf("PSK_%s", randomString(64))
}

func (app *App) createSessionKey(user *user) string {
	ses := &session{
		UUID: generateSessionKey(),
		User: user.ID,
	}
	must(CreateItem(app, ses))
	return ses.UUID
}

func (app *App) getUserIDFromSession(sessionID string) int64 {
	id := app.getSessionCacheUserID(sessionID)
	if id > 0 {
		return id
	}

	ses := Query[session](app).Is("isdeleted", false).Is("uuid", sessionID).First()
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
