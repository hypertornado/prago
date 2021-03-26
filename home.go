package prago

import (
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func (app *App) initHome() {
	app.Action("").Permission(loggedPermission).Name(messages.GetNameFunction("admin_signpost")).Template("admin_home_navigation").DataSource(app.getHomeData)
}

type homeData struct {
	Name string
	URL  string

	Count int64

	Actions []buttonData
}

func (app *App) getHomeData(request *Request) interface{} {
	ret := []homeData{}

	//app.Notification("hello world " + time.Now().Format(time.RFC3339Nano)).Flash(request)
	//app.Notification("hello world 2 " + time.Now().Format(time.RFC3339Nano)).Flash(request)

	for _, resource := range app.getSortedResources(request.user.Locale) {
		if app.authorize(request.user, resource.canView) {
			item := homeData{
				Name: resource.name(request.user.Locale),
				URL:  resource.getURL(""),
			}
			item.Actions = resource.getResourceActionsButtonData(request.user, app)
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
