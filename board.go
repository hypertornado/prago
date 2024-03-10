package prago

import (
	"fmt"
	"time"
)

var sysadminBoard *Board

type Board struct {
	app            *App
	action         *Action
	parentBoard    *Board
	parentResource *Resource

	MainDashboard *Dashboard

	dashboardGroups []*Dashboard
}

func (app *App) initBoard() {
	app.MainBoard = newBoard(app, "").
		Name(messages.GetNameFunction("admin_signpost")).
		Icon(iconSignpost)
	app.MainBoard.action.parentBoard = app.MainBoard
	app.dashboardTableMap = make(map[string]*DashboardTable)
	app.dashboardFigureMap = make(map[string]*DashboardFigure)
	sysadminBoard = app.NewBoard("sysadmin-board").Name(unlocalized("Sysadmin")).Icon("glyphicons-basic-501-server.svg")

	sysadminGroup := sysadminBoard.Dashboard(unlocalized("Sysadmin"))
	sysadminGroup.Figure(unlocalized("Úpravy"), "sysadmin").Value(func(request *Request) int64 {
		c, _ := Query[activityLog](app).Context(request.r.Context()).Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit(unlocalized("/ 24 hodin")).URL("/admin/activitylog").Compare(func(request *Request) int64 {
		c, _ := Query[activityLog](app).Context(request.r.Context()).Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, unlocalized("oproti předchozímu dni"))

	sysadminGroup.Table(func(request *Request) *Table {
		table := app.Table()

		m := map[string]float64{}

		//var graphValues []*graphValue

		for i := 7; i >= 0; i-- {
			c, _ := Query[activityLog](app).Context(request.r.Context()).Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -i-1), time.Now().AddDate(0, 0, -i)).Count()
			//c += 10
			//fmt.Println("XXX", c)
			m[fmt.Sprintf("%d dní", -i)] = float64(c)

			/*graphValues = append(graphValues, &graphValue{
				Name:  fmt.Sprintf("%d dní", -i),
				Value: float64(c),
			})*/
		}

		table.Graph("", GraphDataFromMap(m))
		return table

	}, "sysadmin")
}

func (app *App) NewBoard(url string) *Board {
	board := newBoard(app, url)
	board.parentBoard = app.MainBoard
	return board
}

func newBoard(app *App, url string) *Board {
	ret := &Board{
		app:    app,
		action: app.Action(url),
	}
	ret.action.isPartOfBoard = ret

	ret.action.ui(
		func(request *Request, pd *pageData) {
			pd.BoardView = ret.boardView(request)
		})

	ret.action.permission = loggedPermission
	ret.MainDashboard = &Dashboard{
		name:  unlocalized(""),
		board: ret,
	}
	return ret
}

func (board *Board) Name(name func(string) string) *Board {
	board.action.name = name
	return board
}

func (board *Board) Icon(icon string) *Board {
	board.action.icon = icon
	return board
}

func (board *Board) IsMainBoard() bool {
	return board == board.app.MainBoard
}

func (board *Board) getURL() string {
	if board.parentResource != nil {
		return board.app.getAdminURL(board.parentResource.id)
	}
	return board.app.getAdminURL(board.action.url)
}

func (board *Board) isEmpty(requestContext *menuRequestContext) bool {
	if board.IsMainBoard() {
		return false
	}

	for _, v := range board.dashboardGroups {
		if v.isVisible(requestContext.UserData) {
			return false
		}
	}

	items := board.getMenuItems(requestContext)
	if len(items) > 0 {
		return false
	}
	return true
}
