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
		})
		form.AutosubmitFirstTime = true
		form.AddSubmit("Send")
	}, func(fv FormValidation, request *Request) {

		table := app.Table()

		table.Header("Size", "Access count", "Latest access", "ID")

		stats := app.cache.getStats()

		orderBy := request.Param("order")

		sort.Slice(stats, func(i, j int) bool {

			switch orderBy {
			case "count":
				return stats[i].Count > stats[j].Count
			case "latest":
				return stats[j].LastAccess.Before(stats[i].LastAccess)
			}

			return stats[i].Size > stats[j].Size
		})

		var totalSize int64

		for _, v := range stats {
			totalSize += v.Size
			table.Row(
				Cell(fmt.Sprintf("%s B", humanizeNumber(v.Size))),
				Cell(v.Count),
				Cell(v.LastAccess.Format("2. 1. 2006 15:04:05")),
				Cell(v.ID),
			)
		}

		table.AddFooterText(fmt.Sprintf("Total: %s items, %s B", humanizeNumber(int64(len(stats))), humanizeNumber(totalSize)))

		fv.AfterContent(table.ExecuteHTML())

	}).Name(unlocalized("Cache stats")).Permission("sysadmin").Board(sysadminBoard)
}

type cacheStats struct {
	ID         string
	Size       int64
	Count      int64
	LastAccess time.Time
}
