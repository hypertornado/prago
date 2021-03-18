package prago

import "github.com/gorilla/sessions"

type sessionsManager struct {
	cookieStore *sessions.CookieStore
}

type session struct {
	session *sessions.Session
}

func initRequestWithSession(request *Request) {

}

func (app *App) initSessions() {
	random := app.ConfigurationGetString("random")
	cookieStore := sessions.NewCookieStore([]byte(random))
	app.sessionsManager = &sessionsManager{
		cookieStore: cookieStore,
	}

	app.accessController.addAroundAction(
		func(request Request, next func()) {
			session, err := cookieStore.Get(request.Request(), app.codeName)

			if err != nil {
				app.Log().Println("Session not valid")
				request.Response().Header().Set("Set-Cookie", app.codeName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
				panic(err)
			}

			flashes := session.Flashes()
			if len(flashes) > 0 {
				request.SetData("flash_messages", flashes)
				must(session.Save(request.Request(), request.Response()))
			}

			request.SetData("session", session)
			next()
		},
	)
}
