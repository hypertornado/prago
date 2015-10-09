package prago

func MiddlewareInitSession(request Request) {
	if request.IsProcessed() || request.App().SessionStore() == nil {
		return
	}

	sessionName := "pragoSession"

	session, err := request.App().SessionStore().Get(request.Request(), sessionName)
	if err != nil {
		request.Log().Errorln("Session not valid")
		request.Response().Header().Set("Set-Cookie", sessionName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
		if err != nil {
			panic(err)
		}
	}

	request.SetData("session", session)
}

func MiddlewareSaveSession(request Request) {
	if request.IsProcessed() || request.App().SessionStore() == nil {
		return
	}

	request.Session().Save(request.Request(), request.Response())
}
