package prago

import "github.com/gorilla/sessions"

type sessionsManager struct {
	cookieStore *sessions.CookieStore
}

type requestSession struct {
	session *sessions.Session
	dirty   bool
}

func (rs *requestSession) setValue(key string, value interface{}) {
	rs.session.Values[key] = value
	rs.dirty = true
}

const userIDSessionName = "user_id"

func (request *Request) logInUser(user *user) {
	request.session.setValue(userIDSessionName, user.ID)
}

func (request *Request) logOutUser() {
	delete(request.session.session.Values, userIDSessionName)
	request.session.dirty = true
}

func (request *Request) writeSessionIfDirty() {
	if request.session != nil && request.session.dirty {
		must(request.session.session.Save(request.Request(), request.Response()))
	}
}

//AddFlashMessage adds flash message to request
func (request Request) AddFlashMessage(message string) {
	request.session.session.AddFlash(message)
	request.session.dirty = true
}

func initUserFromSession(request *Request) {
	userID, ok := request.session.session.Values[userIDSessionName].(int64)
	if !ok {
		return
	}

	var user user
	err := request.app.Query().WhereIs("id", userID).Get(&user)
	if err != nil {
		return
	}
	request.user = &user
}

func initRequestWithSession(request *Request, next func()) {
	session, err := request.app.sessionsManager.cookieStore.Get(request.Request(), request.app.codeName)
	if err != nil {
		request.app.Log().Println("Session not valid")
		request.Response().Header().Set("Set-Cookie", request.app.codeName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
		panic(err)
	}

	request.session = &requestSession{
		session: session,
	}
	initUserFromSession(request)

	flashes := session.Flashes()
	if len(flashes) > 0 {
		request.SetData("flash_messages", flashes)
		request.session.dirty = true
	}
	next()
}

func (app *App) initSessions() {
	random := app.ConfigurationGetString("random")
	cookieStore := sessions.NewCookieStore([]byte(random))
	app.sessionsManager = &sessionsManager{
		cookieStore: cookieStore,
	}

	app.accessController.addAroundAction(initRequestWithSession)
}
