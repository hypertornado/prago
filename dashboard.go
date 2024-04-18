package prago

type Dashboard struct {
	board   *Board
	name    func(string) string
	tasks   []*Task
	figures []*dashboardFigure
	tables  []*dashboardTable
}

func (app *App) initDashboard() {

	app.API("dashboard-table").Method("GET").Permission(loggedPermission).Handler(
		func(request *Request) {
			uuid := request.Param("uuid")
			table, err := app.getDashboardTableData(request, uuid)
			must(err)
			err = request.app.templates.templates.ExecuteTemplate(request.Response(), "form_table", table.templateData())
			must(err)
		},
	)

	app.API("dashboard-figure").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("uuid")
			figure, err := app.getDashboardFigureData(request, uuid)
			must(err)
			return figure
		},
	)
}

func (board *Board) Dashboard(name func(string) string) *Dashboard {
	group := &Dashboard{
		board: board,
		name:  name,
	}
	board.dashboardGroups = append(board.dashboardGroups, group)
	return group
}

func (group *Dashboard) isVisible(userData UserData) bool {
	for _, v := range group.figures {
		if userData.Authorize(v.permission) {
			return true
		}
	}

	for _, v := range group.tables {
		if userData.Authorize(v.permission) {
			return true
		}
	}

	for _, v := range group.tasks {
		if userData.Authorize(v.permission) {
			return true
		}
	}

	return false
}
