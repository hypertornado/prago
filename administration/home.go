package administration

import (
	"github.com/hypertornado/prago"
)

type homeData struct {
	Name string
	URL  string

	Count int64

	Actions []buttonData
}

func (admin *Administration) getHomeData(request prago.Request) (ret []homeData) {
	user := GetUser(request)

	for _, resource := range admin.Resources {
		if admin.Authorize(user, resource.CanView) {
			item := homeData{
				Name: resource.HumanName(user.Locale),
				URL:  resource.GetURL(""),
			}
			item.Actions = resource.getResourceActionsButtonData(user, admin)
			ret = append(ret, item)
		}
	}
	return
}
