package prago

func MiddlewareInitSession(request Request) {
	if request.IsProcessed() || request.App().SessionStore() == nil {
		return
	}
	_, r := request.HttpIO()

	sessionName := "pragoSession"

	session, err := request.App().SessionStore().Get(r, sessionName)
	if err != nil {
		request.Log().Errorln("Session not valid")
		w, _ := request.HttpIO()
		w.Header().Set("Set-Cookie", sessionName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
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
	w, r := request.HttpIO()

	request.Session().Save(r, w)
}
