package prago

import (
	"fmt"
	"sort"
	"time"
)

func (app *App) initCacheStats() {

	ActionForm(app, "prago-cache-stats", func(form *Form, request *Request) {
		form.AddSelect("order", "Order By", [][2]string{
			{"size", "Size"},
			{"count", "Access count"},
			{"latest", "Latest access"},
			{"updated", "Latest updated at"},
			{"reload", "Reload duration"},
		})
		form.AutosubmitFirstTime = true
		form.AddSubmit("Send")
	}, func(fv FormValidation, request *Request) {

		table := app.Table()

		table.Header("Size", "Access count", "Latest access", "Last updated at", "Reload duration", "ID")

		stats := app.cache.getStats()

		orderBy := request.Param("order")

		sort.Slice(stats, func(i, j int) bool {

			switch orderBy {
			case "count":
				return stats[i].Count > stats[j].Count
			case "latest":
				return stats[j].LastAccess.Before(stats[i].LastAccess)
			case "updated":
				return stats[j].LastUpdatedAt.Before(stats[i].LastUpdatedAt)
			case "reload":
				return stats[j].ReloadDuration < stats[i].ReloadDuration
			}

			return stats[i].Size > stats[j].Size
		})

		var totalSize int64

		for _, v := range stats {
			totalSize += v.Size
			table.Row(
				Cell(fmt.Sprintf("%s B", humanizeNumber(v.Size))).Nowrap(),
				Cell(v.Count).Nowrap(),
				Cell(v.LastAccess.Format("2. 1. 2006 15:04:05")).Nowrap(),
				Cell(v.LastUpdatedAt.Format("2. 1. 2006 15:04:05")).Nowrap(),
				Cell(v.ReloadDuration.String()).Nowrap(),
				Cell(v.ID),
			)
		}

		table.AddFooterText(fmt.Sprintf("Total: %s items, %s B", humanizeNumber(int64(len(stats))), humanizeNumber(totalSize)))

		fv.AfterContent(table.ExecuteHTML())

	}).Name(unlocalized("Cache stats")).Permission("sysadmin").Board(sysadminBoard)
}

type cacheStats struct {
	ID             string
	Size           int64
	Count          int64
	LastAccess     time.Time
	LastUpdatedAt  time.Time
	ReloadDuration time.Duration
}

func (cache *cache) numberOfItems() (ret int64) {
	cache.items.Range(func(k, v interface{}) bool {
		ret++
		return true
	})
	return ret
}
