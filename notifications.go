package prago

import (
	"time"

	"github.com/hypertornado/prago/utils"
)

type Notification struct {
	ID              int64
	UUID            string `prago-type:"text"`
	Name            string `prago-type:"text"`
	Description     string `prago-type:"text"`
	NotificationTyp string
	IsDismissed     bool
	User            int64 `prago-type:"relation"`
	CreatedAt       time.Time
	UpdatedAt       time.Time `prago-view:"sysadmin"`
}

func initNotificationResource(resource *Resource) {

	resource.App.AdminController.Get(resource.App.GetAdminURL("_api/notifications"), func(request Request) {
		user := request.GetUser()
		notifications, err := resource.App.getNotificationViews(user)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(notifications)
	})

	resource.App.AdminController.Delete(resource.App.GetAdminURL("_api/notification/:uuid"), func(request Request) {
		uuid := request.Params().Get("uuid")
		if uuid == "" {
			panic("wrong length of uuid param")
		}
		var notification Notification
		err := resource.App.Query().WhereIs("uuid", request.Params().Get("uuid")).Get(&notification)
		must(err)
		notification.IsDismissed = true
		must(resource.App.Save(&notification))
		request.RenderJSON(true)
	})
}

type notification struct {
	app         *App
	user        User
	name        string
	description string
	typ         string
}

func (app *App) Notification(user User, name string) *notification {
	return &notification{
		app:  app,
		user: user,
		name: name,
	}
}

func (n *notification) SetDescription(description string) *notification {
	n.description = description
	return n
}

func (n *notification) SetTypeSuccess() *notification {
	n.typ = "success"
	return n
}

func (n *notification) SetTypeFail() *notification {
	n.typ = "fail"
	return n
}

func (n *notification) Create() error {
	item := Notification{
		UUID:            utils.RandomString(10),
		Name:            n.name,
		Description:     n.description,
		NotificationTyp: n.typ,
		User:            n.user.ID,
	}
	return n.app.Create(&item)
}
