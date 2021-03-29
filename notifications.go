package prago

import (
	"sync"
)

//Notification represents user notification
/*type Notification struct {
	ID              int64
	UUID            string `prago-type:"text"`
	Name            string `prago-type:"text"`
	Description     string `prago-type:"text"`
	NotificationTyp string
	IsDismissed     bool
	User            int64 `prago-type:"relation"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}*/

type notificationCenter struct {
	app             *App
	mutex           *sync.RWMutex
	notificationMap map[string]*Notification
}

func (nc *notificationCenter) add(notification *Notification) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	nc.notificationMap[notification.uuid] = notification
}

func (nc *notificationCenter) get(uuid string) *Notification {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	return nc.notificationMap[uuid]
}

func (nc *notificationCenter) delete(uuid string) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	delete(nc.notificationMap, uuid)
}

func (app *App) initNotifications() {
	app.notificationCenter = &notificationCenter{
		app:             app,
		mutex:           &sync.RWMutex{},
		notificationMap: make(map[string]*Notification),
	}

	app.API("notifications").Permission(loggedPermission).Handler(func(request *Request) {
		/*notifications, err := app.getNotificationViews(request.user)
		must(err)
		request.RenderJSON(notifications)*/
	})

	/*.app.adminController.get(resource.app.getAdminURL("_api/notifications"), func(request *Request) {
		notifications, err := resource.app.getNotificationViews(request.user)
		must(err)
		request.RenderJSON(notifications)
	})*/

	app.Action("notification/:uuid").Method("DELETE").Permission(loggedPermission).Handler(func(request *Request) {
		/*uuid := request.Params().Get("uuid")
		var notification Notification
		err := resource.app.Query().WhereIs("uuid", request.Params().Get("uuid")).Get(&notification)
		must(err)
		notification.IsDismissed = true
		must(resource.app.Save(&notification))
		request.RenderJSON(true)*/
	})

}

//NotificationItem represents item for notification
type Notification struct {
	uuid        string
	app         *App
	user        *user
	name        string
	description string
	typ         string
}

//Notification creates notification
func (app *App) Notification(name string) *Notification {
	return &Notification{
		uuid: randomString(10),
		app:  app,
		name: name,
	}
}

//SetDescription sets description to notification item
func (n *Notification) SetDescription(description string) *Notification {
	n.description = description
	return n
}

type notificationView struct {
	Name string
}

func (n *Notification) getView() *notificationView {
	return &notificationView{
		Name: n.name,
	}
}

//SetTypeSuccess sets notification item type to success
func (n *Notification) SetTypeSuccess() *Notification {
	n.typ = "success"
	return n
}

//SetTypeFail sets notification item type to fail
func (n *Notification) SetTypeFail() *Notification {
	n.typ = "fail"
	return n
}

func (n *Notification) Push(user *user) error {
	return nil
}

func (n *Notification) Flash(request *Request) error {
	n.app.notificationCenter.add(n)
	request.session.session.AddFlash(n.uuid)
	request.session.dirty = true
	return nil
}

//Create creates notification
/*func (n *Notification) Create() error {
	item := Notification{
		UUID:            utils.RandomString(10),
		Name:            n.name,
		Description:     n.description,
		NotificationTyp: n.typ,
		User:            n.user.ID,
	}
	return n.app.Create(&item)
}*/
