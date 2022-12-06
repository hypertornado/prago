package prago

type BoardView struct {
	AppName     string
	BoardIcon   string
	BoardName   string
	BoardURL    string
	IsMainBoard bool
	Resources   []menuItem
	UserSection *menuSection

	Dashboards []*DashboardView

	User *user

	TasksName string
	Tasks     *taskViewData
}

type DashboardView struct {
	Name    string
	Figures []*DashboardViewFigure
	Tables  []DashboardViewTable
}

type DashboardViewFigure struct {
	UUID string
	Icon string
	URL  string
	Name string
}

type DashboardViewTable struct {
	UUID string
}

func (board *Board) boardView(request *Request) *BoardView {
	app := board.app
	locale := request.user.Locale
	ret := &BoardView{
		AppName:     request.app.name(request.user.Locale),
		BoardName:   board.action.name(request.user.Locale),
		BoardIcon:   board.action.icon,
		BoardURL:    board.action.getURL(),
		IsMainBoard: board.IsMainBoard(),
		User:        request.user,
	}

	ret.Resources = board.getMenuItems(request)

	if board.IsMainBoard() {
		ret.UserSection = getMenuUserSection(request)
	}

	for _, group := range board.dashboardGroups {
		view := &DashboardView{
			Name: group.name(locale),
		}
		for _, item := range group.figures {
			if request.UserHasPermission(item.permission) {
				view.Figures = append(view.Figures, item.view(app))
			}
		}

		for _, v := range group.tables {
			if request.UserHasPermission(v.permission) {
				view.Tables = append(view.Tables, DashboardViewTable{
					UUID: v.uuid,
				})
			}
		}

		if len(view.Figures) > 0 || len(view.Tables) > 0 {
			ret.Dashboards = append(ret.Dashboards, view)
		}
	}

	if board.IsMainBoard() {
		taskData := GetTaskViewData(request)
		if len(taskData.Tasks) > 0 {
			ret.TasksName = messages.Get(request.user.Locale, "tasks")
			ret.Tasks = &taskData
		}
	}

	return ret
}
