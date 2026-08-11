package prago

import (
	"fmt"
	"html/template"
)

const apiHTTPHeader = "X-API-Key"

func (app *App) initUserAPI() {

	app.ActionUI("_api", func(request *Request) template.HTML {

		table := app.NewTable()
		table.Row(table.Cell("API keys").Header().DescriptionAfter(fmt.Sprintf("Use HTTP header: %s", apiHTTPHeader)))

		usr := request.getUser()
		keys := app.Query[session]().Is("User", usr.ID).Is("IsDeleted", false).Is("IsAPI", true).Order("id").List()
		for _, key := range keys {
			table.Row(Cell(key.UUID).Button(&Button{
				Name:    "Delete",
				Icon:    iconDelete,
				OnClick: template.JS(fmt.Sprintf("popup(\"/admin/_apikeydelete?uuid=%s\")", key.UUID)),
			}))
		}

		table.Row(table.Cell("").Button(&Button{
			Name:    "Generate API Key",
			OnClick: template.JS(fmt.Sprintf("popup(\"/admin/_apikeycreate\")")),
		}))

		return table.ExecuteHTML()

	}).Permission(loggedPermission).Icon("glyphicons-basic-849-computer-network.svg").Name(unlocalized("API")).Board(app.optionsBoard)

	app.PopupForm("_apikeycreate", func(form *Form, request *Request) {
		form.AddSubmit("Generate API key")
	}, func(fv FormValidation, request *Request) {
		usr := request.getUser()
		request.app.createSessionKey(usr, true)
		fv.Data(true)
	}).Name(unlocalized("Generate API key")).Permission(loggedPermission)

	app.PopupForm("_apikeydelete", func(form *Form, request *Request) {
		form.AddTextarea("uuid", "API key").SetValue(request.Param("uuid")).SetReadonly()
		form.AddDeleteSubmit("Delete API key")
	}, func(fv FormValidation, request *Request) {
		err := app.deleteSession(request.Param("uuid"))
		if err != nil {
			fv.AddError(err.Error())
			return
		}

		request.AddFlashMessage("API key deleted")

		fv.Data(true)
	}).Name(unlocalized("Delete API key")).Permission(loggedPermission)

}
