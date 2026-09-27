package prago

import (
	"encoding/json"
	"time"
)

type pragocache struct {
	ID            int64 `prago-order-desc:"true"`
	Name          string
	Value         string `prago-type:"text"`
	ValidDuration int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (app *App) initCachePermanent() {

	app.NewResource[pragocache]().Board(sysadminBoard).Name(Unlocalized("Cache - permanent"), Unlocalized("Cache - permanent")).PermissionView("sysadmin")

}

func (app *App) CachedPermanent[T any](name string, duration time.Duration, createFn func() T) T {

	var ret T

	items := app.Query[pragocache]().Is("name", name).OrderDesc("id").List()

	for _, item := range items {
		if time.Now().Before(item.UpdatedAt.Add(time.Duration(item.ValidDuration))) {
			err := json.Unmarshal([]byte(item.Value), &ret)
			if err == nil {
				return ret
			} else {
				app.Log().Errorf("cant unmarshal cachedpermanent: %s", err)
			}
		} else {
			app.Delete[cacheItem](item.ID)
		}
	}

	ret = createFn()

	value, err := json.Marshal(ret)
	if err != nil {
		panic("cant marshal cachedpermanent: " + err.Error())
	}

	cacheItem := pragocache{
		Name:          name,
		Value:         string(value),
		ValidDuration: int64(duration),
	}

	Must(app.Create(&cacheItem))

	return ret
}
