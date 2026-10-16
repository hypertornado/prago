package prago

import (
	"github.com/hypertornado/prago/pragoai"
)

const claudeAPIKeyID = "claude_api_key"
const claudeAPIAdminKeyID = "claude_api_admin_key"

func (app *App) initAIClient() {
	app.AI = pragoai.NewClaudeClient(app.mustGetSetting(claudeAPIKeyID), app.mustGetSetting(claudeAPIAdminKeyID))
}

var aiBoard *Board

func (app *App) initAI() {

	aiBoard = app.MainBoard.AddChild("_ai", Unlocalized("AI"), iconAI)

	app.NewSetting(claudeAPIKeyID, "sysadmin").ValueChangeCallback(func() {
		app.initAIClient()
	})

	app.NewSetting(claudeAPIAdminKeyID, "sysadmin").ValueChangeCallback(func() {
		app.initAIClient()
	})

	app.initAIClient()
	app.initAIUsage()
	app.initAIConsole()
	app.initAIMCP()
}
