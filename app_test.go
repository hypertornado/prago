package prago

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp(t *testing.T) {
	app := NewApp("prago_test", "v1")
	err := app.initMiddlewares()
	if err != nil {
		t.Fatal(err)
	}

	app.route(get, "/h", app.MainController(), func(request Request) {
		request.SetData("body", []byte("hello"))
	})

	app.route(get, "*some", app.MainController(), func(request Request) {
		s := request.Params().Get("some")
		request.SetData("body", []byte("star "+s))
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/h", nil)
	handleRequest(w, r, app)
	if w.Body.String() != "hello" {
		t.Error(w.Body.String())
	}

	r, _ = http.NewRequest("GET", "/h/", nil)
	handleRequest(w, r, app)
	if w.Body.String() != "hello" {
		t.Error(w.Body.String())
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/b/abc", nil)
	handleRequest(w, r, app)
	if w.Body.String() != "star /b/abc" {
		t.Error(w.Body.String())
	}

}
