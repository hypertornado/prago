package prago

import (
	"sync"
	"time"
)

const staleInterval = 30 * time.Minute

//const staleInterval = 1 * time.Second

type cache struct {
	items map[string]*cacheItem
	mutex *sync.RWMutex
}

type cacheItem struct {
	updatedAt time.Time
	updating  bool
	value     interface{}
	createFn  func() interface{}
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

	val := ci.createFn()

	ci.mutex.Lock()
	ci.value = val
	ci.updatedAt = time.Now()
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

func (c *cache) putItem(name string, createFn func() interface{}) *cacheItem {
	item := &cacheItem{
		updatedAt: time.Now(),
		value:     createFn(),
		createFn:  createFn,
		mutex:     &sync.RWMutex{},
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[name] = item
	return item
}

func loadCache[T any](c *cache, name string, createFn func() T) T {
	/*
		cacheStart := time.Now()
		defer func() {
			durationSince := time.Since(cacheStart)
			if durationSince.Milliseconds() > 0 {
				fmt.Printf("cache '%s', took %s\n", name, durationSince.String())
			}
		}()
	*/

	fn := func() interface{} {
		return createFn()
	}

	item := c.getItem(name)
	if item == nil {
		item := c.putItem(name, fn)
		return item.getValue().(T)
	}

	if item.isStale() {
		//fmt.Println("reloading", name)
		go func() {
			item.reloadValue()
		}()
		return item.getValue().(T)
	}
	return item.getValue().(T)
}

func Cached[T any](app *App, name string, createFn func() T) chan T {
	ret := make(chan T)
	go func() {
		val := loadCache(app.cache, name, createFn)
		ret <- val
	}()
	return ret
}

func (app *App) ClearCache() {
	app.cache.clear()
}

func (c *cache) forceLoad(cacheName string, createFn func() interface{}) interface{} {
	item := c.putItem(cacheName, createFn)
	return item.getValue()
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = map[string]*cacheItem{}
}
