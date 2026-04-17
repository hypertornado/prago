package prago

import (
	"sort"
	"time"
)

func (app *App) initSessionsCache() {
	app.sessionsCacheMap = map[string]int64{}
	app.sessionsCacheLogMap = map[string]*sessionCacheLog{}

	go func() {
		for {
			time.Sleep(30 * time.Second)
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
	sessionID := request.getLoginSessionID()
	if sessionID == "" {
		return
	}
	userAgent := request.Request().UserAgent()
	ipAddress := request.Request().Header.Get("X-Forwarded-For")
	app := request.app
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
		app.Log().Errorf("Can't find session to persist: %s", ses.UUID)
		return
	}
	//TODO: continue when have partial save
	return
}

func (app *App) cleanSessionCache() (ret []*sessionCacheLog) {
	app.sessionsCacheLogMutex.Lock()
	defer app.sessionsCacheLogMutex.Unlock()
	for _, v := range app.sessionsCacheLogMap {
		ret = append(ret, v)
	}
	app.sessionsCacheLogMap = map[string]*sessionCacheLog{}
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
