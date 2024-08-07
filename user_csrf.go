package prago

import (
	"crypto/md5"
	"fmt"
	"io"
)

func (app *App) GenerateCSRFToken(userID int64) string {
	randomness := app.mustGetSetting("random")
	if len(randomness) <= 0 {
		panic("randomness too short")
	}

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%s", userID, randomness))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (request *Request) csrfToken() string {
	return request.app.GenerateCSRFToken(request.UserID())
}

func validateCSRF(request *Request) {
	token := request.csrfToken()
	if len(token) == 0 {
		panic("token not set")
	}
	paramsToken := request.Param("_csrfToken")
	if paramsToken != token {
		panic("Wrong CSRF token")
	}
}
