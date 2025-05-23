package prago

type Dashboard struct {
	board     *Board
	name      func(string) string
	tasks     []*Task
	figures   []*dashboardFigure
	tables    []*dashboardTable
	timelines []*Timeline
}

func (app *App) initDashboard() {

	app.API("dashboard-table").Method("GET").Permission(loggedPermission).Handler(
		func(request *Request) {
			uuid := request.Param("uuid")
			table, err := app.getDashboardTableData(request, uuid)
			must(err)
			err = request.app.adminTemplates.templates.ExecuteTemplate(request.Response(), "form_table", table.templateData())
			must(err)
		},
	)

	app.API("dashboard-figure").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("uuid")
			figure, err := app.getDashboardFigureData(request, uuid)
			if err != nil {
				if err == cantFindFigureError {
					request.WriteJSON(404, "can't find figure")
					return nil
				}
				panic(err)
			}
			return figure
		},
	)

	app.initTimeline()
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

	for _, v := range group.timelines {
		if userData.Authorize(v.permission) {
			return true
		}
	}

	return false
}
