package prago

import (
	"time"
)

type Home struct {
	Sections []*HomeSection
}

type HomeSection struct {
	Name     string
	Items    []*HomeItem
	Tasks    *taskViewData
	UUID     string
	HasTable bool
	//Table    *Table
}

type HomeItem struct {
	Icon        string
	URL         string
	Name        string
	Value       string
	Description string
	IsGreen     bool
	IsRed       bool
}

func (app *App) initHome() {
	app.
		Action("").
		Permission(loggedPermission).
		Name(messages.GetNameFunction("admin_signpost")).
		Template("admin_home_navigation").
		Icon("glyphicons-basic-697-directions-sign.svg").
		DataSource(app.getHomeData)

	sysadminGroup := app.DashboardGroup(unlocalized("Sysadmin"))
	sysadminGroup.Item("Úpravy", "sysadmin").Value(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit("/ 24 hodin").URL("/admin/activitylog").Compare(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, "oproti předchozímu dni")

}

func (app *App) getHomeData(request *Request) interface{} {
	locale := request.user.Locale
	home := &Home{}

	mainMenu := app.getMainMenu(request)

	for _, section := range mainMenu.Sections {
		homeSection := &HomeSection{
			Name: section.Name,
			//UUID: section.UUID,
		}
		home.Sections = append(home.Sections, homeSection)
		for _, item := range section.Items {
			homeItem := &HomeItem{
				Icon: item.Icon,
				Name: item.Name,
				URL:  item.URL,
			}
			homeSection.Items = append(homeSection.Items, homeItem)
		}

	}

	for _, group := range app.dashboardGroups {
		homeSection := &HomeSection{
			Name: group.name(locale),
			UUID: group.uuid,
		}
		for _, item := range group.items {
			if request.UserHasPermission(item.permission) {
				homeSection.Items = append(homeSection.Items, item.homeItem(app))
			}
		}

		if group.table != nil && request.UserHasPermission(group.table.permission) {
			homeSection.HasTable = true
			//homeSection.Table = group.getTable()
		}

		if len(homeSection.Items) > 0 || homeSection.HasTable {
			home.Sections = append(home.Sections, homeSection)
		}
	}

	taskData := GetTaskViewData(request)
	if len(taskData.Tasks) > 0 {
		taskSection := &HomeSection{
			Name:  messages.Get(request.user.Locale, "tasks"),
			Tasks: &taskData,
		}
		home.Sections = append(home.Sections, taskSection)
	}

	return home

}
