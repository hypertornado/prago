package prago

type boardView struct {
	AppName     string
	BoardIcon   string
	BoardName   string
	BoardURL    string
	IsMainBoard bool
	Resources   []*menuItem

	MainDashboard *dashboardView

	Dashboards []*dashboardView

	Role string

	TasksName string
}

type dashboardView struct {
	Name    string
	Tasks   []taskView
	Figures []*dashboardViewFigure
	Tables  []dashboardViewTable
}

type dashboardViewFigure struct {
	UUID               string
	Icon               string
	URL                string
	Name               string
	RefreshTimeSeconds int64
}

type dashboardViewTable struct {
	UUID               string
	RefreshTimeSeconds int64
}

func (board *Board) boardView(request *Request) *boardView {
	locale := request.getUser().Locale

	var boardName, boardIcon, boardURL string
	if board.action != nil {
		boardName = board.action.name(locale)
		if board.isMainBoard() {
			boardName = messages.GetNameFunction("admin_signpost_long", board.app.name(locale))(locale)
		}

		boardIcon = board.action.icon
		boardURL = board.action.getURL()
	}

	if board.parentResource != nil {
		boardName = board.parentResource.pluralName(locale)
		boardIcon = board.parentResource.icon
	}

	ret := &boardView{
		AppName:     board.app.name(request.Locale()),
		BoardName:   boardName,
		BoardIcon:   boardIcon,
		BoardURL:    boardURL,
		IsMainBoard: board.isMainBoard(),
		Role:        request.role(),
	}

	ret.Resources = board.getMenuItems(getMenuRequestContextFromRequest(request, nil))

	ret.MainDashboard = board.mainDashboard.view(request)

	for _, dashboard := range board.dashboardGroups {
		view := dashboard.view(request)
		if view != nil {
			ret.Dashboards = append(ret.Dashboards, view)
		}
	}

	return ret
}

func (dashboard *Dashboard) view(request *Request) *dashboardView {
	if !dashboard.isVisible(request) {
		return nil
	}

	view := &dashboardView{
		Name: dashboard.name(request.Locale()),
	}

	userID := request.UserID()
	csrfToken := request.app.GenerateCSRFToken(userID)
	view.Tasks = dashboard.getTasks(request, csrfToken)

	for _, item := range dashboard.figures {
		if request.Authorize(item.permission) {
			view.Figures = append(view.Figures, item.view(request))
		}
	}

	for _, v := range dashboard.tables {
		if request.Authorize(v.permission) {
			view.Tables = append(view.Tables, dashboardViewTable{
				UUID:               v.uuid,
				RefreshTimeSeconds: v.refreshTimeSeconds,
			})
		}
	}

	return view
}
