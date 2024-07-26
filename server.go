package prago

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

func (app *App) listenAndServe(port int) error {
	app.port = port
	app.Log().Printf("Server started: port=%d, pid=%d, developmentMode=%v\n", port, os.Getpid(), app.developmentMode)

	if !app.developmentMode {
		file, err := os.OpenFile(app.dotPath()+"/prago.log",
			os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
		must(err)
		app.logger.SetOutput(file)
	}

	return (&http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(port),
		Handler:        server{*app},
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}).ListenAndServe()
}

type server struct {
	app App
}

// TODO: remove after fixed tests in lazensky
func (app *App) NewServer() server {
	return server{*app}
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.app.serveHTTP(w, r)
}

func (app *App) serveHTTP(w http.ResponseWriter, r *http.Request) {
	request := &Request{
		uuid:       randomString(10),
		receivedAt: time.Now(),
		w:          w,
		r:          r,
		app:        app,

		ResponseStatus: 200,
	}
	w.Header().Set("X-Prago-Request", request.uuid)

	defer func() {
		if recoveryData := recover(); recoveryData != nil {
			app.recoveryFunction(request, recoveryData)
		}
	}()

	defer func() {
		request.writeAfterLog()
	}()

	if request.removeTrailingSlash() {
		return
	}

	if request.serveStatic() {
		return
	}

	if app.router.process(request) {
		return
	}

	request.Response().WriteHeader(http.StatusNotFound)
	request.Response().Write([]byte("404 — page not found (prago framework)"))
}
