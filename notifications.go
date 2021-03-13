package prago

import (
	"time"

	"github.com/hypertornado/prago/utils"
)

//Notification represents user notification
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

	resource.app.AdminController.Get(resource.app.GetAdminURL("_api/notifications"), func(request Request) {
		user := request.GetUser()
		notifications, err := resource.app.getNotificationViews(user)
		must(err)
		request.RenderJSON(notifications)
	})

	resource.app.AdminController.Delete(resource.app.GetAdminURL("_api/notification/:uuid"), func(request Request) {
		uuid := request.Params().Get("uuid")
		if uuid == "" {
			panic("wrong length of uuid param")
		}
		var notification Notification
		err := resource.app.Query().WhereIs("uuid", request.Params().Get("uuid")).Get(&notification)
		must(err)
		notification.IsDismissed = true
		must(resource.app.Save(&notification))
		request.RenderJSON(true)
	})
}

//NotificationItem represents item for notification
type NotificationItem struct {
	app         *App
	user        User
	name        string
	description string
	typ         string
}

//Notification creates notification
func (app *App) Notification(user User, name string) *NotificationItem {
	return &NotificationItem{
		app:  app,
		user: user,
		name: name,
	}
}

//SetDescription sets description to notification item
func (n *NotificationItem) SetDescription(description string) *NotificationItem {
	n.description = description
	return n
}

//SetTypeSuccess sets notification item type to success
func (n *NotificationItem) SetTypeSuccess() *NotificationItem {
	n.typ = "success"
	return n
}

//SetTypeFail sets notification item type to fail
func (n *NotificationItem) SetTypeFail() *NotificationItem {
	n.typ = "fail"
	return n
}

//Create creates notification
func (n *NotificationItem) Create() error {
	item := Notification{
		UUID:            utils.RandomString(10),
		Name:            n.name,
		Description:     n.description,
		NotificationTyp: n.typ,
		User:            n.user.ID,
	}
	return n.app.Create(&item)
}
