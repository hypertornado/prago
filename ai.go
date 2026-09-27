package prago

import (
	"github.com/hypertornado/prago/pragoai"
)

const claudAPIKeyID = "claude_api_key"
const claudAPIAdminKeyID = "claude_api_admin_key"

func (app *App) initAIClient() {
	app.AI = pragoai.NewClaudeClient(app.mustGetSetting(claudAPIKeyID), app.mustGetSetting(claudAPIAdminKeyID))
}

var aiBoard *Board

func (app *App) initAI() {

	aiBoard = app.MainBoard.AddChild("_ai", Unlocalized("AI"), iconAI)

	app.NewSetting(claudAPIKeyID, "sysadmin").ValueChangeCallback(func() {
		app.initAIClient()
	})

	app.NewSetting(claudAPIAdminKeyID, "sysadmin").ValueChangeCallback(func() {
		app.initAIClient()
	})

	app.initAIClient()
	app.initAIUsage()
	app.initAIConsole()

}
