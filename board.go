package prago

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
