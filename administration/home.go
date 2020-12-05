package administration

import (
	"sort"

	"github.com/hypertornado/prago"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type homeData struct {
	Name string
	URL  string

	Count int64

	Actions []buttonData
}

func (admin *Administration) getHomeData(request prago.Request) (ret []homeData) {
	user := GetUser(request)

	for _, resource := range admin.getSortedResources(user.Locale) {
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

func (admin *Administration) getSortedResources(locale string) (ret []*Resource) {
	collator := collate.New(language.Czech)

	ret = admin.resources
	sort.SliceStable(ret, func(i, j int) bool {
		a := ret[i]
		b := ret[j]

		if collator.CompareString(a.HumanName(locale), b.HumanName(locale)) <= 0 {
			return true
		}
		return false
	})
	return
}
