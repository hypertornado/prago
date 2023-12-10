package prago

import (
	"sync"
)

func (app *App) initUserDataCache() {
	app.userDataCache = make(map[int64]*userData)
	app.userDataCacheMutex = &sync.RWMutex{}
}

func (app *App) userDataCacheGet(id int64) *userData {
	app.userDataCacheMutex.RLock()
	ret := app.userDataCache[id]
	app.userDataCacheMutex.RUnlock()

	if ret != nil {
		return ret
	}

	user := Query[user](app).ID(id)
	if user == nil {
		return nil
	}

	ret = app.newUserData(user)

	app.userDataCacheMutex.Lock()
	defer app.userDataCacheMutex.Unlock()

	app.userDataCache[id] = ret
	return ret
}

func (app *App) userDataCacheDelete(id int64) {
	app.userDataCacheMutex.Lock()
	defer app.userDataCacheMutex.Unlock()
	delete(app.userDataCache, id)
}

func (app *App) userDataCacheDeleteAll() {
	app.userDataCacheMutex.Lock()
	defer app.userDataCacheMutex.Unlock()
	clear(app.userDataCache)
}

type UserData interface {
	Name() string
	//Email() string
	Locale() string
	Authorize(Permission) bool
	//Role() string
	//CSRFToken() string
}

type userData struct {
	name   string
	role   string
	locale string
	app    *App
}

func (app *App) newUserData(user *user) *userData {
	return &userData{
		name:   user.LongName(),
		role:   user.Role,
		locale: user.Locale,
		app:    app,
	}
}

func (d *userData) Name() string {
	return d.name
}

func (d *userData) Locale() string {
	return d.locale
}

func (d *userData) Authorize(permission Permission) bool {
	return d.app.authorize(true, d.role, permission)
}
