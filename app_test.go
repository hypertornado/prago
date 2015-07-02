package prago

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp(t *testing.T) {
	test := NewTest(t)
	app := NewApp()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/h", nil)

	app.Route(GET, "/h", app.MainController(), func(request Request) {
		request.SetData("body", []byte("hello"))
	})

	handleRequest(w, r, app)
	test.EqualString(w.Body.String(), "hello")

}
