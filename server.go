package prago

import (
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
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

	server := &http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(port),
		Handler:        server{*app},
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	if app.serverSetup != nil {
		app.serverSetup(server)
	}

	if app.InitializationAfterServerStart != nil {
		app.InitializationAfterServerStart()
	}

	return server.ListenAndServe()
}

type server struct {
	app App
}

// TODO: remove after fixed tests in lazensky
func (app *App) NewServer() server {
	return server{*app}
}

func (app *App) AddServerSetup(fn func(*http.Server)) {
	app.serverSetup = fn
}

var currentRequestCounter atomic.Int64
var totalRequestCounter atomic.Int64

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentRequestCounter.Add(1)
	totalRequestCounter.Add(1)
	defer func() {
		currentRequestCounter.Add(-1)
	}()

	s.app.serveHTTP(w, r)
}

func (app *App) serveHTTP(w http.ResponseWriter, r *http.Request) {

	if app.exiting {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("503 — unavailable, shutting down"))
		return
	}

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
		if app.afterRequestServedHandler != nil {
			app.afterRequestServedHandler(request)
		}
	}()

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
