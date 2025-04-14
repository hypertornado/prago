package prago

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

var pprofToken = randomString(50)

const pprofPrefix = "/_pprof"

func (app *App) initPprof() {

	for k, v := range map[string]func(http.ResponseWriter, *http.Request){
		pprofPrefix:              pprof.Index,
		pprofPrefix + "/cmdline": pprof.Cmdline,
		pprofPrefix + "/profile": pprof.Profile,
		pprofPrefix + "/symbol":  pprof.Symbol,
		pprofPrefix + "/trace":   pprof.Trace,
	} {
		app.Handle("GET", k, func(request *Request) {
			if request.Param("token") != pprofToken {
				panic("invalid pprof token")
			}
			v(request.Response(), request.Request())
		})
	}
}

func (app *App) getPprofProfilePath() string {
	return fmt.Sprintf("go tool pprof -seconds 5 -svg %s%s/profile?token=%s > flamegraph.svg", app.BaseURL(), pprofPrefix, pprofToken)
}
