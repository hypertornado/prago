package prago

import (
	"time"
)

func cacheReloader(cache *cache) {

	for {
		var oldestItem *cacheItem
		cache.items.Range(func(key, value any) bool {
			v := value.(*cacheItem)

			if oldestItem == nil || v.updatedAt.Get().Before(oldestItem.updatedAt.Get()) {
				oldestItem = v
			}
			return true
		})

		if oldestItem != nil && oldestItem.isStale() {
			oldestItem.reloadValue(cache)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
