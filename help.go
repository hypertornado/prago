package prago

import (
	"fmt"
	"html/template"
)

var helpBoard *Board

func (app *App) AddHelp(url string, permission Permission, name func(string) string, icon string, content func(request *Request) template.HTML) {

	if helpBoard == nil {
		helpBoard = app.optionsBoard.AddChild("help", Unlocalized("Nápověda"), "glyphicons-basic-196-circle-empty-info.svg")
	}

	if icon == "" {
		icon = "glyphicons-basic-196-circle-empty-info.svg"
	}

	app.ActionUI("help/"+url, content).
		Name(name).
		Icon(icon).
		Permission(permission).
		Board(helpBoard)
}

func (app *App) AddHelpMarkdown(url string, permission Permission, name string, icon string, markdownText string) {
	app.AddHelp(url, permission, Unlocalized(name), icon, func(request *Request) template.HTML {
		ret, err := markdownToHTML(markdownText)
		Must(err)

		ret = template.HTML(fmt.Sprintf("<div class=\"box\"><h1>%s</h1>%s</div>", name, ret))

		return ret
	})

}
