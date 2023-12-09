package prago

type BoardView struct {
	AppName     string
	BoardIcon   string
	BoardName   string
	BoardURL    string
	IsMainBoard bool
	Resources   []*menuItem

	MainDashboard *DashboardView

	Dashboards []*DashboardView

	Role string

	TasksName string
}

type DashboardView struct {
	Name    string
	Tasks   []taskView
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
	locale := request.getUser().Locale

	var boardName, boardIcon, boardURL string
	if board.action != nil {
		boardName = board.action.name(locale)
		if board.IsMainBoard() {
			boardName = messages.GetNameFunction("admin_signpost_long", board.app.name(locale))(locale)
		}

		boardIcon = board.action.icon
		boardURL = board.action.getURL()
	}

	if board.parentResource != nil {
		boardName = board.parentResource.pluralName(locale)
		boardIcon = board.parentResource.icon
	}

	ret := &BoardView{
		AppName:     board.app.name(request.Locale()),
		BoardName:   boardName,
		BoardIcon:   boardIcon,
		BoardURL:    boardURL,
		IsMainBoard: board.IsMainBoard(),
		Role:        request.role(),
	}

	ret.Resources = board.getMenuItems(request, nil)

	ret.MainDashboard = board.MainDashboard.view(request)

	for _, dashboard := range board.dashboardGroups {
		view := dashboard.view(request)
		if view != nil {
			ret.Dashboards = append(ret.Dashboards, view)
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

	userID := request.UserID()
	csrfToken := request.app.generateCSRFToken(userID)
	view.Tasks = dashboard.getTasks(userID, request, csrfToken)

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

	return view
}
