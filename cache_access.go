package prago

import "time"

func (c *cache) markAccess(id string) {
	go func() {
		c.accessMutex.Lock()
		defer c.accessMutex.Unlock()

		c.accessCount[id] += 1
		c.lastAccess[id] = time.Now()
	}()
}

func (c *cache) getStats() (ret []cacheStats) {

	c.accessMutex.Lock()
	defer c.accessMutex.Unlock()

	c.items.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(*cacheItem)
		ret = append(ret, cacheStats{
			ID:         k,
			Size:       v.getJSONSize(),
			Count:      c.accessCount[k],
			LastAccess: c.lastAccess[k],
		})

		return true
	})

	return ret
}
