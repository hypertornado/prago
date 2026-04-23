package prago

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (app *App) getFlashCookieID() string {
	return fmt.Sprintf("prago-flash-%s", app.codeName)
}

func (app *App) getLoginCookieID() string {
	return fmt.Sprintf("prago-login-%s", app.codeName)
}

func (request *Request) setCookie(name, value string) {

	var isSecureCookie bool
	if !request.app.developmentMode {
		isSecureCookie = true
	}

	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureCookie,
		Expires:  time.Now().AddDate(100, 0, 0),
		SameSite: http.SameSiteLaxMode,
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

// UserID returns id of logged in user, returns 0 if no user is logged
func (request *Request) UserID() int64 {
	app := request.app
	sessionID, isAPI := request.getLoginSessionID()
	if sessionID == "" {
		return 0
	}
	ret := app.getUserIDFromSession(sessionID, isAPI)
	if ret == 0 {
		request.deleteCookie(app.getLoginCookieID())
	}
	return ret

}

func (request *Request) getLoginSessionID() (string, bool) {
	app := request.app

	apiHeader := request.r.Header.Get(apiHTTPHeader)

	if strings.HasPrefix(apiHeader, "PAK_") {
		return apiHeader, true
	}

	cookies := request.r.CookiesNamed(app.getLoginCookieID())
	if len(cookies) == 0 {
		return "", false
	}
	return cookies[0].Value, false
}

func (request *Request) logInUser(user *user) {
	app := request.app
	sessionID := app.createSessionKey(user, false)
	request.setCookie(app.getLoginCookieID(), sessionID)
}

func (request *Request) logOutUser() {
	app := request.app
	sessionID, isAPI := request.getLoginSessionID()
	if isAPI {
		panic("can't logout api this way")
	}
	if sessionID != "" {
		must(app.deleteSession(sessionID))
	}
	request.AddFlashMessage("Odhlášení proběhlo v pořádku")
	request.deleteCookie(app.getLoginCookieID())
}
