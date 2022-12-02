package prago

import (
	"time"
)

type Home struct {
	IsMainBoard bool
	Sections    []*HomeSection
}

type HomeSection struct {
	Name     string
	Items    []*HomeItem
	Tasks    *taskViewData
	UUID     string
	HasTable bool
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
	sysadminGroup := sysadminBoard.DashboardGroup(unlocalized("Sysadmin"))
	sysadminGroup.Item("Úpravy", "sysadmin").Value(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit("/ 24 hodin").URL("/admin/activitylog").Compare(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, "oproti předchozímu dni")

}

func (board *Board) homeData(request *Request) *Home {
	app := board.app
	locale := request.user.Locale
	home := &Home{
		IsMainBoard: board.IsMainBoard(),
	}

	mainSection := &HomeSection{}
	items := board.getMenuItems(request)
	for _, item := range items {
		homeItem := &HomeItem{
			Icon: item.Icon,
			Name: item.Name,
			URL:  item.URL,
		}
		mainSection.Items = append(mainSection.Items, homeItem)
	}
	home.Sections = append(home.Sections, mainSection)

	if board.IsMainBoard() {

		userMenuSection := getMenuUserSection(request)
		userSection := &HomeSection{
			Name: userMenuSection.Name,
		}
		for _, item := range userMenuSection.Items {
			homeItem := &HomeItem{
				Icon: item.Icon,
				Name: item.Name,
				URL:  item.URL,
			}
			userSection.Items = append(userSection.Items, homeItem)
		}
		home.Sections = append(home.Sections, userSection)
	}

	for _, group := range board.dashboardGroups {
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

	if board.IsMainBoard() {
		taskData := GetTaskViewData(request)
		if len(taskData.Tasks) > 0 {
			taskSection := &HomeSection{
				Name:  messages.Get(request.user.Locale, "tasks"),
				Tasks: &taskData,
			}
			home.Sections = append(home.Sections, taskSection)
		}
	}

	return home

}
