package prago

import (
	"fmt"
	"time"
)

type session struct {
	ID        int64 `prago-order-desc:"true"`
	UUID      string
	User      int64 `prago-type:"relation"`
	IsDeleted bool
	//IsAPIKey  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (app *App) initSessionsResource() {
	NewResource[session](app).PermissionView("sysadmin").Name(unlocalized("Session"), unlocalized("Sessiony"))
}

func generatSessionKey() string {
	return fmt.Sprintf("PSK_" + randomString(64))
}

func (app *App) createSessionKey(user *user) string {
	ses := &session{
		UUID: generatSessionKey(),
		User: user.ID,
	}
	must(CreateItem[session](app, ses))
	return ses.UUID
}

func (app *App) getUserIDFromSession(sessionID string) int64 {
	ses := Query[session](app).Is("isdeleted", false).Is("uuid", sessionID).First()
	if ses == nil {
		return 0
	}
	return ses.User
}

func (app *App) deleteSession(sessionID string) error {
	ses := Query[session](app).Is("uuid", sessionID).First()
	if ses == nil {
		return fmt.Errorf("Nelze najít session")
	}
	return DeleteItem[session](app, ses.ID)
}
