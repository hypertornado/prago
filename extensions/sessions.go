package extensions

import (
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
)

type Sessions struct {
	cookieStore *sessions.CookieStore
	app         *prago.App
}

func (s *Sessions) Init(app *prago.App) error {
	config, err := app.Config()
	if err != nil {
		return err
	}

	cookieStore := sessions.NewCookieStore([]byte(config["random"]))
	s.cookieStore = cookieStore
	s.app = app
	app.Data()["sessionStore"] = s

	app.MainController().AddAroundAction(s.around)

	return nil
}

func (s *Sessions) around(request prago.Request, next func()) {

	if request.IsProcessed() {
		next()
		return
	}

	sessionName := s.app.Data()["appName"].(string)
	session, err := s.cookieStore.Get(request.Request(), sessionName)
	if err != nil {
		request.Log().Errorln("Session not valid")
		request.Response().Header().Set("Set-Cookie", sessionName+"=; expires=Thu, 01 Jan 1970 00:00:01 GMT;")
		panic(err)
	}

	request.SetData("session", session)
	next()
}
