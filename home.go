package prago

import (
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func (app *App) initHome() {
	app.Action("").Name(messages.GetNameFunction("admin_signpost")).Template("admin_home_navigation").DataSource(app.getHomeData)
}

type homeData struct {
	Name string
	URL  string

	Count int64

	Actions []buttonData
}

func (app *App) getHomeData(request *Request) interface{} {
	user := request.getUser()

	ret := []homeData{}

	for _, resource := range app.getSortedResources(user.Locale) {
		if app.authorize(user, resource.canView) {
			item := homeData{
				Name: resource.name(user.Locale),
				URL:  resource.getURL(""),
			}
			item.Actions = resource.getResourceActionsButtonData(user, app)
			ret = append(ret, item)
		}
	}
	return ret
}

func (app *App) getSortedResources(locale string) (ret []*Resource) {
	collator := collate.New(language.Czech)

	ret = app.resources
	sort.SliceStable(ret, func(i, j int) bool {
		a := ret[i]
		b := ret[j]

		if collator.CompareString(a.name(locale), b.name(locale)) <= 0 {
			return true
		}
		return false
	})
	return
}
