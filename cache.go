package prago

import (
	"context"
	"sync"
	"time"
)

//use https://github.com/sourcegraph/conc

const staleInterval = 10 * time.Minute

//const staleInterval = 1 * time.Second

type cache struct {
	items map[string]*cacheItem
	mutex *sync.RWMutex
}

type cacheItem struct {
	updatedAt time.Time
	updating  bool
	value     interface{}
	createFn  func(context.Context) interface{}
	mutex     *sync.RWMutex
}

func newCache() *cache {
	return &cache{
		items: map[string]*cacheItem{},
		mutex: &sync.RWMutex{},
	}
}

func (ci cacheItem) isStale() bool {
	ci.mutex.RLock()
	defer ci.mutex.RUnlock()
	return ci.updatedAt.Add(staleInterval).Before(time.Now())
}

func (ci cacheItem) getValue() interface{} {
	ci.mutex.RLock()
	defer ci.mutex.RUnlock()
	return ci.value
}

func (ci *cacheItem) reloadValue() {
	ci.mutex.RLock()
	if ci.updating {
		ci.mutex.RUnlock()
		return
	}
	ci.mutex.RUnlock()

	ci.mutex.Lock()
	ci.updating = true
	ci.mutex.Unlock()

	var val interface{}

	var panicked bool

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	val = ci.createFn(context.TODO())

	ci.mutex.Lock()
	if !panicked {
		ci.value = val
		ci.updatedAt = time.Now()
	}
	ci.updating = false
	ci.mutex.Unlock()
}

func (c *cache) getItem(name string) *cacheItem {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, found := c.items[name]
	if !found {
		return nil
	}
	return item
}

func (c *cache) putItem(name string, createFn func(context.Context) interface{}) *cacheItem {
	item := &cacheItem{
		updatedAt: time.Now(),
		value:     createFn(context.Background()),
		createFn:  createFn,
		mutex:     &sync.RWMutex{},
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[name] = item
	return item
}

func loadCache[T any](c *cache, name string, createFn func(context.Context) T) T {
	fn := func(ctx context.Context) interface{} {
		return createFn(ctx)
	}

	item := c.getItem(name)
	if item == nil {
		item := c.putItem(name, fn)
		return item.getValue().(T)
	}

	if item.isStale() {
		go func() {
			item.reloadValue()
		}()
		return item.getValue().(T)
	}
	return item.getValue().(T)
}

func Cached[T any](app *App, name string, createFn func(context.Context) T) chan T {
	ret := make(chan T)
	go func() {
		val := loadCache(app.cache, name, createFn)
		ret <- val
	}()
	return ret
}

func (app *App) ClearCache() {
	app.cache.clear()
	app.userDataCacheDeleteAll()
}

func (c *cache) forceLoad(cacheName string, createFn func(context.Context) interface{}) interface{} {
	item := c.putItem(cacheName, createFn)
	return item.getValue()
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = map[string]*cacheItem{}
}
