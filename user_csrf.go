package prago

import (
	"crypto/md5"
	"fmt"
	"io"
)

//CSRFToken generates csrf token for user
func (user *User) csrfToken(randomness string) string {
	if len(randomness) <= 0 {
		panic("randomness too short")
	}

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%s%s", user.ID, randomness, user.LoggedInTime))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//CSRFToken returns csrf token from request
func csrfToken(request Request) string {
	return request.GetData("_csrfToken").(string)
}

//AddCSRFToken adds csrf token to form
func (form *form) AddCSRFToken(request Request) *form {
	form.CSRFToken = csrfToken(request)
	return form
}

//ValidateCSRF validates csrf token for request
func validateCSRF(request Request) {
	token := csrfToken(request)
	if len(token) == 0 {
		panic("token not set")
	}
	paramsToken := request.Params().Get("_csrfToken")
	if paramsToken != token {
		panic("Wrong CSRF token")
	}
}
