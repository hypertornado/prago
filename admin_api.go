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
			request.WriteJSON(200, markdown.New(markdown.HTML(true), markdown.Breaks(true)).RenderToString(data))
		},
	)

	app.API("relationlist").Method("POST").Permission(loggedPermission).Handler(generateRelationListAPIHandler)

	app.API("resource-item-stats").Permission(loggedPermission).Handler(itemStatsAPIHandler)

	app.API("imagepicker").Permission(app.FilesResource.canView).HandlerJSON(imagePickerAPIHandler)

}
