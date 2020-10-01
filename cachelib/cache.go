package cachelib

import (
	"fmt"
	"sync"
	"time"
)

const staleInterval = 30 * time.Minute

type Cache struct {
	items map[string]*cacheItem
	mutex *sync.RWMutex
}

type cacheItem struct {
	id        string
	updatedAt time.Time
	updating  bool
	value     interface{}
	createFn  func() interface{}
	mutex     *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: map[string]*cacheItem{},
		mutex: &sync.RWMutex{},
	}
}

func (ci cacheItem) isStale() bool {
	ci.mutex.RLock()
	defer ci.mutex.RUnlock()
	if ci.updatedAt.Add(staleInterval).Before(time.Now()) {
		return true
	}
	return false
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

	defer func() {
		ci.updating = false
		ci.mutex.Unlock()
	}()

	val := ci.createFn()

	ci.mutex.Lock()
	ci.value = val
	ci.updatedAt = time.Now()
}

func (c *Cache) getItem(name string) *cacheItem {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, found := c.items[name]
	if !found {
		return nil
	}
	return item
}

func (c *Cache) putItem(name string, createFn func() interface{}) *cacheItem {
	item := &cacheItem{
		id:        name,
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

func (c *Cache) Load(cacheName string, createFn func() interface{}) interface{} {
	item := c.getItem(cacheName)
	if item == nil {
		item := c.putItem(cacheName, createFn)
		return item.getValue()
	}

	if item.isStale() {
		go func() {
			item.reloadValue()
		}()
		return item.getValue()
	}
	return item.getValue()
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = map[string]*cacheItem{}
}

func (c *Cache) print() {
	for k, v := range c.items {
		fmt.Println(k, v)
	}
}
