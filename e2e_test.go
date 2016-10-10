package prago

import (
	"github.com/hypertornado/prago/selenium"
	"net/http"
	"testing"
	"time"
)

var (
	testServerRunning bool
	driver            *selenium.Driver
)

func TestE2E(t *testing.T) {
	prepareE2ETest(t)

	session := driver.NewTestSession(t)
	defer session.Delete()

	session.SetURL("http://localhost:8587/h")
	source := session.GetElementByID("content").Text()
	if source != "hello" {
		t.Fatal(source)
	}

	session.SetURL("http://localhost:8587/h/")
	source = session.GetElementByID("content").Text()
	if source != "hello" {
		t.Fatal(source)
	}

	session.SetURL("http://localhost:8587/b/abc")
	source = session.GetElementByID("content").Text()
	if source != "/b/abc" {
		t.Fatal(session.GetSource())
	}
}

func prepareE2ETest(t *testing.T) {
	if testServerRunning {
		return
	}

	driver = selenium.NewDriver("http://localhost:9515")
	app := NewApp("prago", "1")
	app.Config = &config{map[string]interface{}{}}

	app.AddMiddleware(MiddlewareServer{initE2Etest})
	err := app.initMiddlewares()
	if err != nil {
		t.Fatal(err)
	}
	go app.ListenAndServe(8587, true)

	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, _ := http.Get("http://localhost:8587/test")
		if resp.StatusCode == 200 {
			testServerRunning = true
			return
		}
	}

}

func initE2Etest(app *App) {
	app.MainController().Get("/h", func(request Request) {
		request.Header().Add("Content-type", "text/html")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte("<div id=\"content\">hello</div>"))
		request.SetProcessed()
	})
	app.MainController().Get("*some", func(request Request) {
		request.Header().Add("Content-type", "text/html")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte("<div id=\"content\">" + request.Params().Get("some") + "</div>"))
		request.SetProcessed()
	})
}
