package prago

import (
	"net/http"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Redirect(request Request, urlStr string) {
	request.SetProcessed()
	request.Header().Set("Location", urlStr)
	request.Response().WriteHeader(http.StatusMovedPermanently)
}
