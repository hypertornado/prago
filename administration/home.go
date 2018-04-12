package administration

import (
	"github.com/hypertornado/prago"
)

type homeData struct {
	Name string
	URL  string

	Count int64

	Actions []ButtonData
}

func (a *Administration) getHomeData(request prago.Request) (ret []homeData) {
	user := GetUser(request)
	locale := GetLocale(request)

	for _, resource := range a.Resources {
		if resource.HasView && resource.Authenticate(user) {
			item := homeData{
				Name: resource.Name(locale),
				URL:  resource.GetURL(""),
			}

			if resource.HasView {
				item.Actions = resource.ResourceActionsButtonData(user, a)
			}
			ret = append(ret, item)
		}
	}
	return
}
