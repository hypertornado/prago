package prago

import (
	//"github.com/hypertornado/prago"
	"testing"
)

func TestE2E(t *testing.T) {

	app := NewApp("prago", "1")
	app.Config = &config{map[string]interface{}{}}

	app.AddMiddleware(MiddlewareServer{initE2Etest})
	err := app.initMiddlewares()
	if err != nil {
		t.Fatal(err)
	}
	go app.ListenAndServe(8587, true)

}

func initE2Etest(app *App) {
	app.MainController().Get("/", func(request Request) {
		panic("eee")
	})
}
