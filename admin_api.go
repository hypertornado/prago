package prago

import (
	"io"

	"github.com/golang-commonmark/markdown"
)

func (app *App) initAPI() {
	app.API("markdown").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			data, err := io.ReadAll(request.Request().Body)
			must(err)
			request.RenderJSON(markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data))
		},
	)

	app.API("relationlist").Method("POST").Permission(loggedPermission).Handler(generateRelationListAPIHandler)

}
