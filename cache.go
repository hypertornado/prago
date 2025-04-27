package prago

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

//use https://github.com/sourcegraph/conc

const staleInterval = 10 * time.Minute

type cache struct {
	items sync.Map
	group singleflight.Group

	totalRequests   atomic.Int64
	currentRequests atomic.Int64
	reloadWaiting   atomic.Int64

	//accessMutex *sync.RWMutex
}

type cacheItem struct {
	once           sync.Once
	updatedAt      syncedItem[time.Time]
	lastAccess     syncedItem[time.Time]
	value          syncedItem[any]
	createFn       func() any
	reloadDuration syncedItem[time.Duration]

	accessCount atomic.Int64
}

func newCache() *cache {
	ret := &cache{}
	go cacheReloader(ret)
	return ret
}

func (item *cacheItem) isStale() bool {
	return item.updatedAt.Get().Add(staleInterval).Before(time.Now())
}

func (item *cacheItem) getValue() any {
	return item.value.Get()
}

func (item *cacheItem) reloadValue(cache *cache) {
	cache.reloadWaiting.Add(1)
	defer func() {
		cache.reloadWaiting.Add(-1)
	}()

	var val any
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from cache createFn panic: %v", err)
		}
	}()

	var reloadStart = time.Now()
	val = item.createFn()

	item.value.Set(val)
	item.updatedAt.Set(time.Now())
	item.reloadDuration.Set(time.Now().Sub(reloadStart))
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

//use https://pkg.go.dev/golang.org/x/sync/singleflight

func (cache *cache) putItem(name string, createFn func() any) *cacheItem {

	cache.reloadWaiting.Add(1)
	defer func() {
		cache.reloadWaiting.Add(-1)
	}()

	var reloadStart = time.Now()
	item := &cacheItem{
		updatedAt:  syncedItem[time.Time]{},
		lastAccess: syncedItem[time.Time]{},
		value: syncedItem[any]{
			val: createFn(),
		},
		createFn:       createFn,
		reloadDuration: syncedItem[time.Duration]{},
	}
	item.reloadDuration.Set(time.Now().Sub(reloadStart))

	cache.items.Store(name, item)
	return item
}

func loadCache[T any](cache *cache, name string, createFn func() T) T {

	i, err, _ := cache.group.Do(name, func() (interface{}, error) {
		fn := func() any {
			return createFn()
		}
		item := cache.getItem(name)
		if item == nil {
			item = cache.putItem(name, fn)
		}
		return item, nil
	})

	if err != nil {
		panic(err)
	}

	item := i.(*cacheItem)

	item.accessCount.Add(1)
	item.lastAccess.Set(time.Now())
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
}
