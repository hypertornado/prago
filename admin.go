package prago

import (
	"errors"
	"fmt"

	_ "embed"

	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/messages"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

//go:embed static/public/admin/_static/admin.js
var staticAdminJS []byte

//go:embed static/public/admin/_static/admin.css
var staticAdminCSS []byte

//go:embed static/public/admin/_static/pikaday.js
var staticPikadayJS []byte

//GetURL gets url
func (app App) GetURL(suffix string) string {
	ret := app.prefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//AddAction adds action
func (app *App) AddAction(action Action) {
	bindAction(app, nil, action, false)
	app.rootActions = append(app.rootActions, action)
}

//AddFieldType adds field type
func (app *App) AddFieldType(name string, fieldType FieldType) {
	_, exist := app.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	app.fieldTypes[name] = fieldType
}

//AddJavascript adds javascript
func (app *App) AddJavascript(url string) {
	app.javascripts = append(app.javascripts, url)
}

//AddCSS adds CSS
func (app *App) AddCSS(url string) {
	app.css = append(app.css, url)
}

//AddFlashMessage adds flash message to request
func AddFlashMessage(request Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	must(session.Save(request.Request(), request.Response()))
}

func addCurrentFlashMessage(request Request, message string) {
	data := request.GetData("flash_messages")
	messages, _ := data.([]interface{})
	messages = append(messages, message)
	request.SetData("flash_messages", messages)
}

//GetItemURL gets item url
func (resource Resource) GetItemURL(item interface{}, suffix string) string {
	ret := resource.GetURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func render403(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request Request) {
	request.SetData("message", messages.Messages.Get(getLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}
