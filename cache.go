package prago

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"
)

//use https://github.com/sourcegraph/conc

const staleInterval = 30 * time.Minute

//const staleInterval = 1 * time.Second

type cache struct {
	items sync.Map

	reloadMutex *sync.RWMutex

	accessMutex *sync.RWMutex
	accessCount map[string]int64
	lastAccess  map[string]time.Time
}

type cacheItem struct {
	updatedAt time.Time
	updating  bool
	value     any
	createFn  func() any
	mutex     *sync.RWMutex
}

func newCache() *cache {
	ret := &cache{
		reloadMutex: &sync.RWMutex{},
		accessMutex: &sync.RWMutex{},
		accessCount: map[string]int64{},
		lastAccess:  map[string]time.Time{},
	}
	return ret
}

func (ci cacheItem) isStale() bool {
	ci.mutex.RLock()
	defer ci.mutex.RUnlock()

	//disperse in time
	randomStaleCoeficient := rand.Intn(30*60) * int(time.Second)

	return ci.updatedAt.Add(staleInterval + time.Duration(randomStaleCoeficient)).Before(time.Now())
}

func (ci cacheItem) getJSONSize() int64 {
	val := ci.getValue()
	data, err := json.Marshal(val)
	if err != nil {
		return -1
	}
	return int64(len(data))

}

func (ci cacheItem) getValue() any {
	ci.mutex.RLock()
	defer ci.mutex.RUnlock()
	return ci.value
}

func (ci *cacheItem) reloadValue(c *cache) {

	c.reloadMutex.Lock()
	defer c.reloadMutex.Unlock()

	ci.mutex.RLock()
	if ci.updating {
		ci.mutex.RUnlock()
		return
	}
	ci.mutex.RUnlock()

	ci.mutex.Lock()
	ci.updating = true
	ci.mutex.Unlock()

	var val any

	var panicked bool

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	val = ci.createFn()

	ci.mutex.Lock()
	if !panicked {
		ci.value = val
		ci.updatedAt = time.Now()
	}
	ci.updating = false
	ci.mutex.Unlock()
}

func (c *cache) getItem(name string) *cacheItem {
	item, ok := c.items.Load(name)
	if !ok {
		return nil
	}
	return item.(*cacheItem)
}

func (c *cache) putItem(name string, createFn func() any) *cacheItem {
	item := &cacheItem{
		updatedAt: time.Now(),
		value:     createFn(),
		createFn:  createFn,
		mutex:     &sync.RWMutex{},
	}

	c.items.Store(name, item)
	return item
}

func loadCache[T any](c *cache, name string, createFn func() T) T {
	fn := func() any {
		return createFn()
	}

	item := c.getItem(name)
	if item == nil {
		item := c.putItem(name, fn)
		return item.getValue().(T)
	}

	if item.isStale() {
		go func() {
			item.reloadValue(c)
		}()
	}
	return item.getValue().(T)
}

func Cached[T any](app *App, name string, createFn func() T) chan T {
	ret := make(chan T)
	go func() {
		val := loadCache(app.cache, name, createFn)
		ret <- val
	}()
	app.cache.markAccess(name)
	return ret
}

func (app *App) ClearCache() {
	app.cache.clear()
	app.userDataCacheDeleteAll()
}

func (c *cache) forceLoad(cacheName string, createFn func() any) any {
	item := c.putItem(cacheName, createFn)
	return item.getValue()
}

func (c *cache) clear() {

	c.items.Range(func(key, value any) bool {
		c.items.Delete(key)
		return true
	})

	c.accessMutex.Lock()
	defer c.accessMutex.Unlock()

	c.accessCount = map[string]int64{}
	c.lastAccess = map[string]time.Time{}
}
