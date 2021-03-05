package prago

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	_ "embed"

	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

//go:embed static/public/admin/_static/admin.js
var staticAdminJS []byte

//go:embed static/public/admin/_static/admin.css
var staticAdminCSS []byte

//go:embed static/public/admin/_static/pikaday.js
var staticPikadayJS []byte

//go:embed templates
var templatesFS embed.FS

//GetURL gets url
func (app App) GetURL(suffix string) string {
	ret := app.prefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

//AddAction adds action
func (admin *App) AddAction(action Action) {
	bindAction(admin, nil, action, false)
	admin.rootActions = append(admin.rootActions, action)
}

//AddFieldType adds field type
func (admin *App) AddFieldType(name string, fieldType FieldType) {
	_, exist := admin.fieldTypes[name]
	if exist {
		panic(fmt.Sprintf("field type '%s' already set", name))
	}
	admin.fieldTypes[name] = fieldType
}

//AddJavascript adds javascript
func (admin *App) AddJavascript(url string) {
	admin.javascripts = append(admin.javascripts, url)
}

//AddCSS adds CSS
func (admin *App) AddCSS(url string) {
	admin.css = append(admin.css, url)
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

func (admin *App) getResourceByName(name string) *Resource {
	return admin.resourceNameMap[columnName(name)]
}

func (admin *App) getDB() *sql.DB {
	return admin.db
}

//GetDB gets DB
func (admin *App) GetDB() *sql.DB {
	return admin.getDB()
}

func (admin *App) initAutoRelations() {
	for _, v := range admin.resources {
		v.initAutoRelations()
	}
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

func columnName(fieldName string) string {
	return utils.PrettyURL(fieldName)
}
