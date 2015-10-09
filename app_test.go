package prago

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp(t *testing.T) {
	app := NewApp()

	app.Route(GET, "/h", app.MainController(), func(request Request) {
		request.SetData("body", []byte("hello"))
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/h", nil)
	handleRequest(w, r, app)
	if w.Body.String() != "hello" {
		t.Error(w.Body.String())
	}

}
