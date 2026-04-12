package prago

import (
	"fmt"
	"net/http"
	"time"
)

func (app *App) getFlashCookieID() string {
	return fmt.Sprintf("prago-flash-" + app.codeName)
}

func (app *App) getLoginCookieID() string {
	return fmt.Sprintf("prago-login-" + app.codeName)
}

func (request *Request) setCookie(name, value string) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().AddDate(100, 0, 0),
	}
	http.SetCookie(request.w, &cookie)
}

func (request *Request) deleteCookie(name string) {
	cookie := http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}
	http.SetCookie(request.w, &cookie)
}
