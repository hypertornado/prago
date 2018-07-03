package main

import (
	"fmt"
	"net/http"
	"time"
)

var configMap = map[string]ConfigServer{}

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	for _, v := range config.Servers {
		configMap[v.Name] = v
		server := NewServer(v)
		server.Start()
	}

	/*port, err := Freeport()
	if err != nil {
		panic(err)
	}*/

	err = (&http.Server{
		Addr:           "0.0.0.0:80",
		Handler:        server{},
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}).ListenAndServe()
	if err != nil {
		panic(err)
	}
}

type server struct{}

func (server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hello, %s", r.Host)
}
