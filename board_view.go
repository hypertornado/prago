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

	Role string

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
		AppName:     request.app.name(request.Locale()),
		BoardName:   board.action.name(request.Locale()),
		BoardIcon:   board.action.icon,
		BoardURL:    board.action.getURL(),
		IsMainBoard: board.IsMainBoard(),
		Role:        request.role(),
	}

	ret.Resources, _ = board.getMainItems(request)

	ret.MainDashboard = board.MainDashboard.view(request)

	if board.IsMainBoard() {
		ret.UserSection = getMenuUserSection(request)
	}

	for _, dashboard := range board.dashboardGroups {
		view := dashboard.view(request)
		if view != nil {
			ret.Dashboards = append(ret.Dashboards, view)
		}

	}

	if board.IsMainBoard() {
		taskData := GetTaskViewData(request)
		if len(taskData.Tasks) > 0 {
			ret.TasksName = messages.Get(request.Locale(), "tasks")
			ret.Tasks = &taskData
		}
	}

	return ret
}

func (dashboard *Dashboard) view(request *Request) *DashboardView {
	if !dashboard.isVisible(request) {
		return nil
	}

	view := &DashboardView{
		Name: dashboard.name(request.Locale()),
	}
	for _, item := range dashboard.figures {
		if request.Authorize(item.permission) {
			view.Figures = append(view.Figures, item.view(request))
		}
	}

	for _, v := range dashboard.tables {
		if request.Authorize(v.permission) {
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
