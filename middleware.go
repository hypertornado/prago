package prago

import (
	"net/http"
	"strings"
)

var (
	MiddlewareStatic              = requestMiddleware{requestMiddlewareStatic}
	MiddlewareDispatcher          = requestMiddleware{requestMiddlewareDispatcher}
	MiddlewareRemoveTrailingSlash = requestMiddleware{requestMiddlewareRemoveTrailingSlash}
	MiddlewareParseRequest        = requestMiddleware{requestMiddlewareParseRequest}
	MiddlewareLogBefore           = requestMiddleware{requestMiddlewareLogBefore}
)

type Middleware interface {
	Init(*App) error
}

type requestMiddleware struct {
	rm RequestMiddleware
}

func (m requestMiddleware) Init(app *App) error {
	app.requestMiddlewares = append(app.requestMiddlewares, m.rm)
	return nil
}

func requestMiddlewareRemoveTrailingSlash(p Request, next func()) {
	path := p.Request().URL.Path
	if p.Request().Method == "GET" && len(path) > 1 && path == p.Request().URL.String() && strings.HasSuffix(path, "/") {
		Redirect(p, path[0:len(path)-1])
		p.Response().WriteHeader(http.StatusMovedPermanently)
		p.SetProcessed()
	}
	next()
}

func requestMiddlewareParseRequest(r Request, next func()) {

	if !r.IsProcessed() {
		contentType := r.Request().Header.Get("Content-Type")
		var err error

		if strings.HasPrefix(contentType, "multipart/form-data") {
			err = r.Request().ParseMultipartForm(1000000)
			if err != nil {
				panic(err)
			}

			for k, values := range r.Request().MultipartForm.Value {
				for _, v := range values {
					r.Request().Form.Add(k, v)
				}
			}
		} else {
			err = r.Request().ParseForm()
			if err != nil {
				panic(err)
			}
		}
	}

	next()
}
