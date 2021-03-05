package prago

import (
	"github.com/gorilla/sessions"
)

func createSessionAroundAction(appName, random string) func(Request, func()) {
	cookieStore := sessions.NewCookieStore([]byte(random))
	return func(request Request, next func()) {
		session, err := cookieStore.Get(request.Request(), appName)
		if err != nil {
			request.Log().Println("Session not valid")
			request.Response().Header().Set("Set-Cookie", appName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
			panic(err)
		}

		flashes := session.Flashes()
		if len(flashes) > 0 {
			request.SetData("flash_messages", flashes)
			must(session.Save(request.Request(), request.Response()))
		}

		request.SetData("session", session)
		next()
	}
}
