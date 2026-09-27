package prago

import (
	"sync"
	"time"
)

var userPreferenceMutex sync.RWMutex

type pragouserpreference struct {
	ID        int64 `prago-order-desc:"true"`
	Name      string
	Value     string `prago-type:"text"`
	User      int64  `prago-type:"relation"`
	CreatedAt time.Time
	UpdatedAt time.Time `prago-can-view:"sysadmin"`
}

func initUserPreference(app *App) {
	resource := app.NewResource[pragouserpreference]()
	resource.Name(
		messages.GetNameFunction("User preference"),
		messages.GetNameFunction("User preferences"),
	)
	resource.PermissionView(sysadminPermission)
	resource.Board(sysadminBoard)
}

func (request *Request) GetUserPreference(name string) *string {
	userID := request.UserID()
	if userID == 0 {
		return nil
	}

	userPreferenceMutex.RLock()
	defer userPreferenceMutex.RUnlock()

	preference := request.app.Query[pragouserpreference]().Is("user", userID).Is("name", name).First()
	if preference != nil {
		return &preference.Value
	}
	return nil
}

func (request *Request) SetUserPreference(name string, value string) {
	userID := request.UserID()
	if userID == 0 {
		panic("cant find user for preference")
	}

	userPreferenceMutex.Lock()
	defer userPreferenceMutex.Unlock()

	preference := request.app.Query[pragouserpreference]().Is("user", userID).Is("name", name).First()
	if preference != nil {
		preference.Value = value
		Must(request.UpdateWithLog(preference))
	} else {
		preference = &pragouserpreference{
			Name:  name,
			Value: value,
			User:  request.UserID(),
		}
		Must(request.CreateWithLog(preference))
	}
}

func (request *Request) DeleteUserPreference(name string) {
	userID := request.UserID()
	if userID == 0 {
		panic("cant find user for preference")
	}

	userPreferenceMutex.Lock()
	defer userPreferenceMutex.Unlock()

	preference := request.app.Query[pragouserpreference]().Is("user", userID).Is("name", name).First()
	if preference != nil {
		Must(request.DeleteWithLog(preference))
	}
}
