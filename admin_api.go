package prago

import (
	"io"
)

func (app *App) initAPI() {
	app.NewAPI("markdown").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			data, err := io.ReadAll(request.Request().Body)
			Must(err)
			htmlData, err := markdownToHTML(string(data))
			Must(err)
			request.WriteJSON(200, htmlData)
		},
	)

	app.NewAPI("relationlist").Method("POST").Permission(loggedPermission).Handler(generateRelationListAPIHandler)

	app.NewAPI("resource-item-stats").Permission(loggedPermission).Handler(itemStatsAPIHandler)

	app.NewAPI("_fetch_list_cell_relation").Permission(loggedPermission).Handler(fetchListCellRelationAPIHandler)

}
