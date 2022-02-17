package prago

import (
	"crypto/md5"
	"fmt"
	"io"
)

var csrfRandomness string

func RequestCSRF(request *Request) string {
	if csrfRandomness == "" {
		csrfRandomness = randomString(30)
	}
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", csrfRandomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))
}
