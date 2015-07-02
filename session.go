package prago

func MiddlewareInitSession(request Request) {
	if request.IsProcessed() || request.App().SessionStore() == nil {
		return
	}
	_, r := request.HttpIO()
	session, err := request.App().SessionStore().Get(r, "pragoSession")
	Must(err)

	request.SetData("session", session)
}

func MiddlewareSaveSession(request Request) {
	if request.IsProcessed() || request.App().SessionStore() == nil {
		return
	}
	w, r := request.HttpIO()
	request.Session().Save(r, w)
}
