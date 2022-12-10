package prago

type BoardView struct {
	AppName     string
	BoardIcon   string
	BoardName   string
	BoardURL    string
	IsMainBoard bool
	Resources   []menuItem
	UserSection *menuSection

	MainDashboard *DashboardView

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
	UUID               string
	Icon               string
	URL                string
	Name               string
	RefreshTimeSeconds int64
}

type DashboardViewTable struct {
	UUID               string
	RefreshTimeSeconds int64
}

func (board *Board) boardView(request *Request) *BoardView {
	ret := &BoardView{
		AppName:     request.app.name(request.user.Locale),
		BoardName:   board.action.name(request.user.Locale),
		BoardIcon:   board.action.icon,
		BoardURL:    board.action.getURL(),
		IsMainBoard: board.IsMainBoard(),
		User:        request.user,
	}

	ret.Resources, _ = board.getMainItems(request)

	ret.MainDashboard = board.MainDashboard.view(request)

	if board.IsMainBoard() {
		ret.UserSection = getMenuUserSection(request)
	}

	for _, dashboard := range board.dashboardGroups {
		/*if !group.isVisible(app, request.user) {
			continue
		}

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
					UUID:               v.uuid,
					RefreshTimeSeconds: v.refreshTimeSeconds,
				})
			}
		}

		if len(view.Figures) > 0 || len(view.Tables) > 0 {
			ret.Dashboards = append(ret.Dashboards, view)
		}*/

		view := dashboard.view(request)
		if view != nil {
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

func (dashboard *Dashboard) view(request *Request) *DashboardView {
	app := request.app
	if !dashboard.isVisible(app, request.user) {
		return nil
	}

	view := &DashboardView{
		Name: dashboard.name(request.user.Locale),
	}
	for _, item := range dashboard.figures {
		if request.UserHasPermission(item.permission) {
			view.Figures = append(view.Figures, item.view(app))
		}
	}

	for _, v := range dashboard.tables {
		if request.UserHasPermission(v.permission) {
			view.Tables = append(view.Tables, DashboardViewTable{
				UUID:               v.uuid,
				RefreshTimeSeconds: v.refreshTimeSeconds,
			})
		}
	}

	if len(view.Figures) > 0 || len(view.Tables) > 0 {
		return view

	}
	return nil

}
