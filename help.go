package prago

import (
	"html/template"
)

var helpBoard *Board

func (app *App) Help(url string, name func(string) string, content func(request *Request) template.HTML) {

	if helpBoard == nil {
		helpBoard = app.MainBoard.Child("help", unlocalized("Nápověda"), "glyphicons-basic-196-circle-empty-info.svg")
	}

	ActionUI(app, "help/"+url, content).
		Name(name).
		Icon("glyphicons-basic-196-circle-empty-info.svg").
		Permission(loggedPermission).
		Board(helpBoard)
}
