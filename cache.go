package prago

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

//use https://github.com/sourcegraph/conc

const staleInterval = 30 * time.Minute

type cache struct {
	items sync.Map

	totalRequests   atomic.Int64
	currentRequests atomic.Int64
	reloadWaiting   atomic.Int64

	//reloadingValuesCount sync.

	reloadMutex *sync.RWMutex
	accessMutex *sync.RWMutex
	accessCount map[string]int64
	lastAccess  map[string]time.Time
}

type cacheItem struct {
	updatedAt      time.Time
	updating       bool
	value          any
	createFn       func() any
	reloadDuration time.Duration
	mutex          *sync.RWMutex
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

func (item cacheItem) isStale() bool {
	item.mutex.RLock()
	defer item.mutex.RUnlock()

	//disperse in time
	randomStaleCoeficient := rand.Intn(30*60) * int(time.Second)

	return item.updatedAt.Add(staleInterval + time.Duration(randomStaleCoeficient)).Before(time.Now())
}

func (item cacheItem) getValue() any {
	item.mutex.RLock()
	defer item.mutex.RUnlock()
	return item.value
}

func (item *cacheItem) reloadValue(cache *cache) {
	cache.reloadWaiting.Add(1)
	defer func() {
		cache.reloadWaiting.Add(-1)
	}()

	item.mutex.RLock()
	if item.updating {
		item.mutex.RUnlock()
		return
	}
	item.mutex.RUnlock()

	item.mutex.Lock()
	if item.updating {
		item.mutex.Unlock()
		return
	}
	item.updating = true
	item.mutex.Unlock()

	cache.reloadMutex.Lock()
	defer cache.reloadMutex.Unlock()

	var val any
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from cache createFn panic: %v", err)
			item.mutex.Lock()
			item.updating = false
			item.mutex.Unlock()
		}
	}()

	var reloadStart = time.Now()
	val = item.createFn()
	var reloadEnd = time.Now()

	item.mutex.Lock()
	item.value = val
	item.updatedAt = time.Now()
	item.updating = false
	item.reloadDuration = reloadEnd.Sub(reloadStart)
	item.mutex.Unlock()
}

func (cache *cache) getItem(name string) *cacheItem {
	cache.totalRequests.Add(1)
	cache.currentRequests.Add(1)
	defer func() {
		cache.currentRequests.Add(-1)
	}()

	item, ok := cache.items.Load(name)
	if !ok {
		return nil
	}
	return item.(*cacheItem)
}

func (cache *cache) putItem(name string, createFn func() any) *cacheItem {
	var reloadStart = time.Now()
	item := &cacheItem{
		updatedAt: time.Now(),
		value:     createFn(),
		createFn:  createFn,
		updating:  false,
		mutex:     &sync.RWMutex{},
	}
	item.reloadDuration = time.Now().Sub(reloadStart)

	cache.items.Store(name, item)
	return item
}

func loadCache[T any](cache *cache, name string, createFn func() T) T {
	fn := func() any {
		return createFn()
	}

	item := cache.getItem(name)
	if item == nil {
		item := cache.putItem(name, fn)
		return item.getValue().(T)
	}

	go func() {
		if item.isStale() {
			item.reloadValue(cache)
		}
	}()

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
