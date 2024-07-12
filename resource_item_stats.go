package prago

import (
	"fmt"
)

type itemStat struct {
	id         string
	Name       func(string) string
	Permission Permission
	Handler    func(item any) string
}

func ItemStatistic[T any](app *App, name func(string) string, permission Permission, statHandler func(item *T) string) {
	resource := getResource[T](app)
	resource.itemStats = append(resource.itemStats, &itemStat{
		id:         fmt.Sprintf("_stat-%d", len(resource.itemStats)+1),
		Name:       name,
		Permission: permission,
		Handler: func(item any) string {
			return statHandler(item.(*T))
		},
	})
}

type itemStatResponse struct {
	Value string
}

func itemStatsAPIHandler(request *Request) {
	resource := request.app.resourceNameMap[request.Param("resource_id")]
	if !request.Authorize(resource.canView) {
		panic("not allowed resource")
	}

	var stat *itemStat
	for _, v := range resource.itemStats {
		if v.id == request.Param("stat_id") {
			stat = v
			break
		}
	}
	if !request.Authorize(stat.Permission) {
		panic("not allowed stat")
	}

	item := resource.query(request.r.Context()).ID(request.Param("item_id"))

	request.WriteJSON(200, itemStatResponse{
		Value: stat.Handler(item),
	})

}
