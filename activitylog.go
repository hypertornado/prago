package prago

import (
	"encoding/json"
	"fmt"
	"time"
)

//ActivityLog logs all user admin activity
type ActivityLog struct {
	ID            int64
	ResourceName  string    `prago-preview:"true"`
	ItemID        int64     `prago-preview:"true"`
	ActionType    string    `prago-preview:"true"`
	User          int64     `prago-type:"relation" prago-preview:"true"`
	ContentBefore string    `prago-type:"text"`
	ContentAfter  string    `prago-type:"text"`
	CreatedAt     time.Time `prago-preview:"true"`
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

//ListenActivityLog listens to all changes in app's administration
func (app *App) ListenActivityLog(handler func(logItem ActivityLog)) {
	app.activityListeners = append(app.activityListeners, handler)
}

func (app *App) createActivityLog(log ActivityLog) error {
	err := app.Create(&log)
	if err == nil {
		for _, v := range app.activityListeners {
			v(log)
		}
	}
	return err
}

func (app *App) getHistory(resource *Resource, itemID int64) historyView {
	ret := historyView{}

	q := app.Query()
	if resource != nil {
		q.WhereIs("ResourceName", resource.ID)
	}
	if itemID > 0 {
		q.WhereIs("ItemID", itemID)
	}
	q.Limit(250)
	q.OrderDesc("ID")

	var items []*ActivityLog
	must(q.Get(&items))

	for _, v := range items {
		var username, userurl string

		var user User
		err := app.Query().WhereIs("id", v.User).Get(&user)
		if err == nil {
			username = user.Name
			userurl = app.GetAdminURL(fmt.Sprintf("user/%d", user.ID))
		}

		activityURL := app.GetAdminURL(fmt.Sprintf("activitylog/%d", v.ID))
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
	resource.CanView = permissionSysadmin
	resource.OrderDesc = true
	resource.HumanName = messages.GetNameFunction("admin_history")
}

func (app App) createNewActivityLog(resource Resource, user User, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return app.createActivityLog(ActivityLog{
		ResourceName: resource.ID,
		ItemID:       getItemID(item),
		ActionType:   "new",
		User:         user.ID,
		ContentAfter: string(data),
	})
}

func (app App) createEditActivityLog(resource Resource, user User, itemID int64, before, after []byte) error {
	return app.createActivityLog(ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "edit",
		User:          user.ID,
		ContentBefore: string(before),
		ContentAfter:  string(after),
	})
}

func (app App) createDeleteActivityLog(resource Resource, user User, itemID int64, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return app.createActivityLog(ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "delete",
		User:          user.ID,
		ContentBefore: string(data),
	})
}
