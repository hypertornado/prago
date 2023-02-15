package prago

import (
	"fmt"
	"sync"
)

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

func (nc *notificationCenter) getFromUUID(uuid string) *Notification {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	return nc.notificationMap[uuid]
}

func (nc *notificationCenter) getFromUser(userID int64) (ret []*notificationView) {
	ret = []*notificationView{}
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	for _, v := range nc.notificationMap {
		if v.userID > 0 && userID > 0 && v.userID == userID {
			ret = append(ret, v.getView())
		}
	}
	return
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

	app.API("notifications").Permission(everybodyPermission).Handler(func(request *Request) {
		var notifications []*notificationView = []*notificationView{}
		userID := request.UserID()
		if userID > 0 {
			notifications = app.notificationCenter.getFromUser(userID)
		}
		request.RenderJSON(notifications)
	})

	app.API("notifications").Method("POST").Permission(loggedPermission).Handler(func(request *Request) {
		action := request.Param("action")
		uuid := request.Param("uuid")
		switch action {
		case "delete":
			app.notificationCenter.delete(uuid)
		case "primary":
			notification := app.notificationCenter.getFromUUID(uuid)
			notification.primaryAction.fn()
		case "secondary":
			notification := app.notificationCenter.getFromUUID(uuid)
			notification.secondaryAction.fn()
		default:
			panic("unknown action " + action)
		}
	})
}

// NotificationItem represents item for notification
type Notification struct {
	uuid            string
	app             *App
	userID          int64
	preName         string
	image           string
	url             string
	name            string
	description     string
	primaryAction   *notificationItemAction
	secondaryAction *notificationItemAction
	disableCancel   bool
	style           string
	progress        *notificationProgress
}

type notificationProgress struct {
	Human      string
	Percentage float64
}

// Notification creates notification
func (app *App) Notification(name string) *Notification {
	return &Notification{
		uuid: randomString(10),
		app:  app,
		name: name,
	}
}

// SetDescription sets description to notification item
func (n *Notification) SetDescription(description string) *Notification {
	n.description = description
	return n
}

// SetPreName sets prefix name to notification item
func (n *Notification) SetPreName(preName string) *Notification {
	n.preName = preName
	return n
}

func (n *Notification) SetImage(image string) *Notification {
	n.image = image
	return n
}

func (n *Notification) SetURL(url string) *Notification {
	n.url = url
	return n
}

// SetProgress sets description to notification item
func (n *Notification) SetProgress(progress *float64) *Notification {

	if progress == nil {
		n.progress = nil
	} else {
		if *progress < 0 {
			n.progress = &notificationProgress{
				Human:      "",
				Percentage: -1,
			}
		} else {
			n.progress = &notificationProgress{
				Human:      notificationProgressHuman(*progress),
				Percentage: *progress,
			}
		}
	}
	return n
}

func (n *Notification) SetPrimaryAction(name string, fn func()) *Notification {
	n.primaryAction = &notificationItemAction{
		name: name,
		fn:   fn,
	}
	return n
}

func (n *Notification) SetSecondaryAction(name string, fn func()) *Notification {
	n.secondaryAction = &notificationItemAction{
		name: name,
		fn:   fn,
	}
	return n
}

type notificationItemAction struct {
	name string
	fn   func()
}

type notificationView struct {
	UUID            string
	PreName         string
	Image           string
	URL             string
	Name            string
	Description     string
	PrimaryAction   *string
	SecondaryAction *string
	DisableCancel   bool
	Style           string
	Progress        *notificationProgress
}

func (n *Notification) getView() *notificationView {

	var primaryAction, secondaryAction *string
	if n.primaryAction != nil {
		primaryAction = &n.primaryAction.name
	}

	if n.secondaryAction != nil {
		secondaryAction = &n.secondaryAction.name
	}

	return &notificationView{
		UUID:            n.uuid,
		PreName:         n.preName,
		Image:           n.image,
		URL:             n.url,
		Name:            n.name,
		Description:     n.description,
		PrimaryAction:   primaryAction,
		SecondaryAction: secondaryAction,
		DisableCancel:   n.disableCancel,
		Style:           n.style,
		Progress:        n.progress,
	}
}

// SetTypeSuccess sets notification item type to success
func (n *Notification) SetStyleSuccess() *Notification {
	n.style = "success"
	return n
}

// SetTypeFail sets notification item type to fail
func (n *Notification) SetStyleFail() *Notification {
	n.style = "fail"
	return n
}

func (n *Notification) Push(userID int64) {
	n.userID = userID
	n.app.notificationCenter.add(n)
}

func (n *Notification) Flash(request *Request) error {
	n.app.notificationCenter.add(n)
	request.session.session.AddFlash(n.uuid)
	request.session.dirty = true
	return nil
}

func notificationProgressHuman(in float64) string {
	if in <= 0 {
		return ""
	}
	if in > 1 {
		return ""
	}
	return fmt.Sprintf("%.2f %%", in*100)
}
