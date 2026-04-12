package prago

import (
	"github.com/gorilla/sessions"
)

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

// AddFlashMessage adds flash message to request
func (request *Request) AddFlashMessage(message string) {
	request.app.Notification(message).Flash(request)
}

func (n *Notification) Flash(request *Request) error {
	n.isFlash = true
	n.app.notificationCenter.add(n)
	request.setCookie(n.app.getFlashCookieID(), n.uuid)
	return nil
}

func initRequestWithSession(request *Request) {
	app := request.app
	session, err := request.app.sessionsManager.cookieStore.Get(request.Request(), request.app.codeName)
	if err != nil {
		request.app.Log().Println("Session not valid")
		request.Response().Header().Set("Set-Cookie", request.app.codeName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
		panic(err)
	}

	request.session = &requestSession{
		session: session,
	}

	var notifications []*notificationView

	flashCookies := request.r.CookiesNamed(app.getFlashCookieID())
	for _, cookie := range flashCookies {
		notification := request.app.notificationCenter.getFromUUID(cookie.Value)
		if notification != nil {
			notifications = append(notifications, notification.getView())
			request.deleteCookie(app.getFlashCookieID())
		}

	}
	request.notifications = notifications
}

func (app *App) initSessions() {
	random := app.mustGetSetting("random")
	cookieStore := sessions.NewCookieStore([]byte(random))
	app.sessionsManager = &sessionsManager{
		cookieStore: cookieStore,
	}

	app.mainController.addBeforeAction(initRequestWithSession)
}
