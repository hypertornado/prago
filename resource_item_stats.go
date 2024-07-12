package prago

import "fmt"

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
