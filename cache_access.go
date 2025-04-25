package prago

import (
	"encoding/json"
	"time"
)

func (c *cache) getStats() (ret []cacheStats) {

	c.items.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(*cacheItem)

		var updateAt time.Time
		item := c.getItem(k)
		if item != nil {
			updateAt = item.updatedAt.Get()
		}

		ret = append(ret, cacheStats{
			ID:             k,
			Size:           v.getJSONSize(),
			Count:          v.accessCount.Load(),
			LastAccess:     v.lastAccess.Get(),
			LastUpdatedAt:  updateAt,
			ReloadDuration: v.reloadDuration.Get(),
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
