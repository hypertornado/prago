package prago

import (
	"encoding/json"
	"time"
)

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

		var updateAt time.Time
		item := c.getItem(k)
		if item != nil {
			updateAt = item.updatedAt
		}

		ret = append(ret, cacheStats{
			ID:             k,
			Size:           v.getJSONSize(),
			Count:          c.accessCount[k],
			LastAccess:     c.lastAccess[k],
			LastUpdatedAt:  updateAt,
			ReloadDuration: v.reloadDuration,
		})

		return true
	})

	return ret
}

func (ci cacheItem) getJSONSize() int64 {
	val := ci.getValue()
	data, err := json.Marshal(val)
	if err != nil {
		return -1
	}
	return int64(len(data))
}
