package prago

import (
	"time"
)

func cacheReloader(cache *cache) {

	for {
		var oldestItem *cacheItem
		//var oldestItemID string

		cache.items.Range(func(key, value any) bool {
			v := value.(*cacheItem)

			if oldestItem == nil || v.updatedAt.Before(oldestItem.updatedAt) {
				oldestItem = v
				//oldestItemID = key.(string)
			}
			return true
		})

		if oldestItem != nil && oldestItem.isStale() {
			//fmt.Println("reloading stale item " + oldestItemID)
			oldestItem.reloadValue(cache)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
