package prago

import (
	"context"
	"sort"
	"sync"
	"time"
)

func (app *App) initSessionsCache() {
	app.sessionsCacheMutex = &sync.RWMutex{}
	app.sessionsCacheMap = map[string]int64{}

	app.sessionsCacheLogMutex = &sync.Mutex{}
	app.sessionsCacheLogMap = map[string]*sessionCacheLog{}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			persistSessionItems := app.cleanSessionCache()
			sort.Slice(persistSessionItems, func(i, j int) bool {
				if persistSessionItems[i].AccessAt.Before(persistSessionItems[j].AccessAt) {
					return true
				}
				return false
			})
			for _, item := range persistSessionItems {
				app.persistSessionCacheItem(item)
			}
		}
	}()
}

type sessionCacheLog struct {
	SessionUUID string
	AccessAt    time.Time
	UserAgent   string
	IPAddress   string
}

func (request *Request) writeToSessionCache() {
	app := request.app
	sessionID, _ := request.getLoginSessionID()
	if sessionID == "" {
		return
	}
	userAgent := request.Request().UserAgent()
	ipAddress := request.Request().Header.Get("X-Forwarded-For")
	go func() {
		app.sessionsCacheLogMutex.Lock()
		defer app.sessionsCacheLogMutex.Unlock()
		log := &sessionCacheLog{
			SessionUUID: sessionID,
			AccessAt:    time.Now(),
			UserAgent:   userAgent,
			IPAddress:   ipAddress,
		}
		app.sessionsCacheLogMap[sessionID] = log
	}()
}

func (app *App) persistSessionCacheItem(item *sessionCacheLog) {

	ses := Query[session](app).Is("uuid", item.SessionUUID).First()
	if ses == nil {
		app.Log().Errorf("Can't find session to persist: %s", item.SessionUUID)
		return
	}

	err := app.sessionsResource.saveItem(context.Background(), &session{
		ID:         ses.ID,
		UserAgent:  item.UserAgent,
		IPAddress:  item.IPAddress,
		LastAccess: item.AccessAt,
	}, map[string]bool{
		"UserAgent":  true,
		"IPAddress":  true,
		"LastAccess": true,
	}, false)

	if err != nil {
		app.Log().Errorf("Can't persist %s: %s", ses.UUID, err)
		return
	}

	usr := Query[user](app).ID(ses.User)
	if usr == nil {
		app.Log().Errorf("Can't find user to persist %d", ses.User)
		return
	}

	err = app.UsersResource.saveItem(context.Background(), &user{
		ID:         usr.ID,
		UserAgent:  item.UserAgent,
		IPAddress:  item.IPAddress,
		LastAccess: item.AccessAt,
	}, map[string]bool{
		"UserAgent":  true,
		"IPAddress":  true,
		"LastAccess": true,
	}, false)

	if err != nil {
		app.Log().Errorf("Can't persist user %d: %s", usr.ID, err)
		return
	}
}

func (app *App) cleanSessionCache() (ret []*sessionCacheLog) {
	app.sessionsCacheLogMutex.Lock()
	defer app.sessionsCacheLogMutex.Unlock()
	for _, v := range app.sessionsCacheLogMap {
		ret = append(ret, v)
	}
	for k := range app.sessionsCacheLogMap {
		delete(app.sessionsCacheLogMap, k)
	}
	return
}

func (app *App) getSessionCacheUserID(sessionID string) int64 {
	app.sessionsCacheMutex.RLock()
	defer app.sessionsCacheMutex.RUnlock()
	return app.sessionsCacheMap[sessionID]
}

func (app *App) setSessionCacheUserID(sessionID string, userID int64) {
	app.sessionsCacheMutex.Lock()
	defer app.sessionsCacheMutex.Unlock()
	app.sessionsCacheMap[sessionID] = userID
}

func (app *App) deleteSessionCacheUserID(sessionID string) {
	app.sessionsCacheMutex.Lock()
	defer app.sessionsCacheMutex.Unlock()
	delete(app.sessionsCacheMap, sessionID)
}
