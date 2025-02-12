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

	mainDashboard *Dashboard

	dashboardGroups []*Dashboard
}

func (app *App) initBoard() {
	app.MainBoard = newBoard(app, "")
	app.MainBoard.action.name = messages.GetNameFunction("admin_signpost")
	app.MainBoard.action.icon = iconSignpost
	app.MainBoard.action.parentBoard = app.MainBoard
	app.dashboardTableMap = make(map[string]*dashboardTable)
	app.dashboardFigureMap = make(map[string]*dashboardFigure)
	app.dashboardTimelineMap = make(map[string]*Timeline)
	sysadminBoard = app.MainBoard.Child("sysadmin-board", unlocalized("Sysadmin"), "glyphicons-basic-501-server.svg")

	sysadminGroup := sysadminBoard.Dashboard(unlocalized("Sysadmin"))
	sysadminGroup.Figure(unlocalized("Úpravy"), "sysadmin").Value(func(request *Request) int64 {
		c, _ := Query[activityLog](app).Context(request.r.Context()).Where("createdat >= ?", time.Now().AddDate(0, 0, -1)).Count()
		return c
	}).Unit(unlocalized("/ 24 hodin")).URL("/admin/activitylog").Compare(func(request *Request) int64 {
		c, _ := Query[activityLog](app).Context(request.r.Context()).Where("createdat >= ? and createdat <= ?", time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -1)).Count()
		return c
	}, unlocalized("oproti předchozímu dni"))

	tl := sysadminGroup.Timeline(unlocalized("Úpravy"), "sysadmin")

	tl.DataSource(func(request *TimelineDataRequest) float64 {
		c, _ := Query[activityLog](app).Context(request.Context).Where("createdat >= ? and createdat < ?", request.From, request.To).Count()
		return float64(c)
	}).Name(unlocalized("Úpravy")).Stringer(func(f float64) string {
		return fmt.Sprintf("%v editací", f)
	})

}

func (parent *Board) Child(url string, name func(string) string, icon string) *Board {
	app := parent.app
	board := newBoard(app, url)
	board.parentBoard = parent
	board.action.name = name
	board.action.icon = icon
	return board
}

func newBoard(app *App, url string) *Board {
	ret := &Board{
		app:    app,
		action: ActionPlain(app, url, nil),
	}
	ret.action.isPartOfBoard = ret

	ret.action.ui(
		func(request *Request, pd *pageData) {
			pd.BoardView = ret.boardView(request)
		})

	ret.action.permission = loggedPermission
	ret.mainDashboard = &Dashboard{
		name:  unlocalized(""),
		board: ret,
	}
	return ret
}

func (board *Board) isMainBoard() bool {
	return board == board.app.MainBoard
}

func (board *Board) getURL() string {
	if board.parentResource != nil {
		return board.app.getAdminURL(board.parentResource.id)
	}
	return board.app.getAdminURL(board.action.url)
}

func (board *Board) isEmpty(requestContext *menuRequestContext) bool {
	if board.isMainBoard() {
		return false
	}

	for _, v := range board.dashboardGroups {
		if v.isVisible(requestContext.UserData) {
			return false
		}
	}

	items := board.getMenuItems(requestContext)
	return len(items) <= 0
}
