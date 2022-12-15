package prago

import (
	"time"
)

var sysadminBoard *Board

type Board struct {
	app         *App
	action      *Action
	parentBoard *Board

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
	sysadminBoard = app.NewBoard("sysadmin-board").Name(unlocalized("Sysadmin"))

	app.NewBoard("empty-board")

	sysadminGroup := sysadminBoard.Dashboard(unlocalized("Sysadmin"))
	sysadminGroup.Figure(unlocalized("Úpravy"), "sysadmin").Value(func(request *Request) int64 {
		c, _ := app.activityLogResource.Query(request.r.Context()).Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit(unlocalized("/ 24 hodin")).URL("/admin/activitylog").Compare(func(request *Request) int64 {
		c, _ := app.activityLogResource.Query(request.r.Context()).Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, unlocalized("oproti předchozímu dni"))
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
	ret.action.template = "board"
	ret.action.dataSource = func(request *Request) interface{} {
		return ret.boardView(request)
	}
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

func (board *Board) isEmpty(request *Request) bool {
	if board.IsMainBoard() {
		return false
	}

	for _, v := range board.dashboardGroups {
		if v.isVisible(request.app, request.user) {
			return false
		}
	}

	items, _ := board.getItems(request, false)
	if len(items) > 0 {
		return false
	}

	return true
}
