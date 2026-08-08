package prago

type TemporaryCacheData[T any] struct {
	app   *App
	cache map[int64]*T
}

func TemporaryCache[T any](app *App) *TemporaryCacheData[T] {
	return &TemporaryCacheData[T]{
		app:   app,
		cache: make(map[int64]*T),
	}
}

func (tc TemporaryCacheData[T]) GetItemByID(id int64) *T {
	ret, ok := tc.cache[id]
	if ok {
		return ret
	}
	item := tc.app.Query[T]().ID(id)
	tc.cache[id] = item
	return item

}
