package admin

import (
	"github.com/hypertornado/prago"
)

type HomeData struct {
	Name string
	URL  string

	Count int64

	Actions []ButtonData
}

func (a *Admin) GetHomeData(request prago.Request) (ret []HomeData) {
	user := GetUser(request)
	locale := GetLocale(request)

	for _, resource := range a.Resources {
		if resource.HasView && resource.Authenticate(user) {
			item := HomeData{
				Name: resource.Name(locale),
				URL:  a.Prefix + "/" + resource.ID,
			}

			if resource.HasModel {
				var ifaceItem interface{}
				resource.newItem(&ifaceItem)
				item.Count, _ = a.Query().Count(ifaceItem)
				item.Actions = resource.ResourceActionsButtonData(user, a)
			}
			ret = append(ret, item)
		}
	}
	return
}
