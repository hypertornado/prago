package prago

import (
	"context"
	"time"
)

var sysadminBoard *Board

type Board struct {
	app         *App
	action      *Action
	parentBoard *Board

	dashboardGroups []*Dashboard
}

func (app *App) initBoard() {
	app.MainBoard = newBoard(app, "").
		Name(messages.GetNameFunction("admin_signpost")).
		Icon(iconSignpost)
	app.MainBoard.action.parentBoard = app.MainBoard
	app.dashboardTableMap = make(map[string]*dashboardTable)
	app.dashboardFigureMap = make(map[string]*DashboardFigure)
	sysadminBoard = app.NewBoard("sysadmin-board").Name(unlocalized("Sysadmin"))

	sysadminGroup := sysadminBoard.Dashboard(unlocalized("Sysadmin"))
	sysadminGroup.Figure("Úpravy", "sysadmin").Value(func(ctx context.Context) int64 {
		c, _ := app.activityLogResource.Query(ctx).Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit("/ 24 hodin").URL("/admin/activitylog").Compare(func(ctx context.Context) int64 {
		c, _ := app.activityLogResource.Query(ctx).Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, "oproti předchozímu dni")
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
