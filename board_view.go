package prago

import (
	"time"
)

type BoardView struct {
	AppName     string
	BoardIcon   string
	BoardName   string
	BoardURL    string
	IsMainBoard bool
	Resources   []menuItem
	UserSection *menuSection

	Sections []*BoardViewSection

	User *user
}

type BoardViewSection struct {
	Name  string
	Items []*BoardViewItem
	Tasks *taskViewData
	//UUID   string
	Tables []BoardViewTable
}

type BoardViewTable struct {
	UUID string
}

type BoardViewItem struct {
	Icon        string
	URL         string
	Name        string
	Value       string
	Description string
	IsGreen     bool
	IsRed       bool
}

func (app *App) initHome() {
	sysadminGroup := sysadminBoard.Dashboard(unlocalized("Sysadmin"))
	sysadminGroup.Item("Úpravy", "sysadmin").Value(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit("/ 24 hodin").URL("/admin/activitylog").Compare(func() int64 {
		c, _ := app.activityLogResource.Query().Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, "oproti předchozímu dni")

}

func (board *Board) boardView(request *Request) *BoardView {
	app := board.app
	locale := request.user.Locale
	home := &BoardView{
		AppName:     request.app.name(request.user.Locale),
		BoardName:   board.action.name(request.user.Locale),
		BoardIcon:   board.action.icon,
		BoardURL:    board.action.getURL(),
		IsMainBoard: board.IsMainBoard(),
		User:        request.user,
	}

	home.Resources = board.getMenuItems(request)

	if board.IsMainBoard() {
		home.UserSection = getMenuUserSection(request)
	}

	for _, group := range board.dashboardGroups {
		homeSection := &BoardViewSection{
			Name: group.name(locale),
		}
		for _, item := range group.figures {
			if request.UserHasPermission(item.permission) {
				homeSection.Items = append(homeSection.Items, item.homeItem(app))
			}
		}

		for _, v := range group.tables {
			if request.UserHasPermission(v.permission) {
				homeSection.Tables = append(homeSection.Tables, BoardViewTable{
					UUID: v.uuid,
				})
			}
		}

		if len(homeSection.Items) > 0 || len(homeSection.Tables) > 0 {
			home.Sections = append(home.Sections, homeSection)
		}
	}

	if board.IsMainBoard() {
		taskData := GetTaskViewData(request)
		if len(taskData.Tasks) > 0 {
			taskSection := &BoardViewSection{
				Name:  messages.Get(request.user.Locale, "tasks"),
				Tasks: &taskData,
			}
			home.Sections = append(home.Sections, taskSection)
		}
	}

	return home

}
