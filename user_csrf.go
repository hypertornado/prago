package prago

import (
	"crypto/md5"
	"fmt"
	"io"
)

func (app *App) generateCSRFToken(user *User) string {
	randomness := app.ConfigurationGetString("random")
	if len(randomness) <= 0 {
		panic("randomness too short")
	}

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%s%s", user.ID, randomness, user.LoggedInTime))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (request *Request) csrfToken() string {
	return request.app.generateCSRFToken(request.user)
}

func (form *form) AddCSRFToken(request *Request) *form {
	form.CSRFToken = request.csrfToken()
	return form
}

func validateCSRF(request *Request) {
	token := request.csrfToken()
	if len(token) == 0 {
		panic("token not set")
	}
	paramsToken := request.Params().Get("_csrfToken")
	if paramsToken != token {
		panic("Wrong CSRF token")
	}
}
