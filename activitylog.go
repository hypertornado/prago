package prago

import (
	"encoding/json"
	"fmt"
	"time"
)

type activityLog struct {
	ID            int64
	ResourceName  string    `prago-preview:"true"`
	ItemID        int64     `prago-preview:"true"`
	ActionType    string    `prago-preview:"true"`
	User          int64     `prago-type:"relation" prago-preview:"true"`
	ContentBefore string    `prago-type:"text"`
	ContentAfter  string    `prago-type:"text"`
	CreatedAt     time.Time `prago-preview:"true"`
}

type ActivityType int

const (
	ActivityCreate ActivityType = iota
	ActivityEdit
	ActivityDelete
)

func (t ActivityType) string() string {
	switch t {
	case ActivityCreate:
		return "new"
	case ActivityEdit:
		return "edit"
	case ActivityDelete:
		return "delete"
	default:
		return ""
	}
}

type Activity struct {
	ID         int64
	ResourceID string
	User       int64
	Typ        ActivityType
}

func (al activityLog) activity() Activity {

	var typ ActivityType
	switch al.ActionType {
	case "new":
		typ = ActivityCreate
	case "edit":
		typ = ActivityEdit
	case "delete":
		typ = ActivityDelete
	}

	return Activity{
		ID:         al.ItemID,
		ResourceID: al.ResourceName,
		User:       al.User,
		Typ:        typ,
	}
}

type historyView struct {
	Items []historyItemView
}

type historyItemView struct {
	ID          int64
	ActionType  string
	ActivityURL string
	ItemName    string
	ItemURL     string
	UserName    string
	UserURL     string
	CreatedAt   string
}

//ListenActivity listens to all changes in app's administration
func (app *App) ListenActivity(handler func(logItem Activity)) {
	app.activityListeners = append(app.activityListeners, handler)
}

func (app *App) createActivityLog(log activityLog) error {
	err := app.Create(&log)
	if err == nil {
		for _, v := range app.activityListeners {
			v(log.activity())
		}
	}
	return err
}

func (app *App) getHistory(resource *Resource, itemID int64) historyView {
	ret := historyView{}

	q := app.Query()
	if resource != nil {
		q.WhereIs("ResourceName", resource.id)
	}
	if itemID > 0 {
		q.WhereIs("ItemID", itemID)
	}
	q.Limit(250)
	q.OrderDesc("ID")

	var items []*activityLog
	must(q.Get(&items))

	for _, v := range items {
		var username, userurl string

		var user user
		err := app.Query().WhereIs("id", v.User).Get(&user)
		if err == nil {
			username = user.Name
			userurl = app.getAdminURL(fmt.Sprintf("user/%d", user.ID))
		}

		activityURL := app.getAdminURL(fmt.Sprintf("activitylog/%d", v.ID))
		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ID)

		ret.Items = append(ret.Items, historyItemView{
			ID:          v.ID,
			ActivityURL: activityURL,
			ActionType:  v.ActionType,
			ItemName:    itemName,
			ItemURL:     resource.getURL(fmt.Sprintf("%d", v.ItemID)),
			UserName:    username,
			UserURL:     userurl,
			CreatedAt:   messages.Timestamp(user.Locale, v.CreatedAt, true),
		})
	}
	return ret
}

func initActivityLog(resource *Resource) {
	resource.canView = Permission(sysadminRoleName)
	resource.orderDesc = true
	resource.name = messages.GetNameFunction("admin_history")
}

func (app App) createNewActivityLog(resource Resource, user *user, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return app.createActivityLog(activityLog{
		ResourceName: resource.id,
		ItemID:       getItemID(item),
		ActionType:   "new",
		User:         user.ID,
		ContentAfter: string(data),
	})
}

func (app App) createEditActivityLog(resource Resource, user *user, itemID int64, before, after []byte) error {
	return app.createActivityLog(activityLog{
		ResourceName:  resource.id,
		ItemID:        itemID,
		ActionType:    "edit",
		User:          user.ID,
		ContentBefore: string(before),
		ContentAfter:  string(after),
	})
}

func (app App) createDeleteActivityLog(resource Resource, user *user, itemID int64, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return app.createActivityLog(activityLog{
		ResourceName:  resource.id,
		ItemID:        itemID,
		ActionType:    "delete",
		User:          user.ID,
		ContentBefore: string(data),
	})
}
