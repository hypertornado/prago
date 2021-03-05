package prago

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hypertornado/prago/messages"
)

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

func (admin *App) ListenActivityLog(handler func(logItem ActivityLog)) {
	admin.activityListeners = append(admin.activityListeners, handler)
}

func (admin *App) createActivityLog(log ActivityLog) error {
	err := admin.Create(&log)
	if err == nil {
		for _, v := range admin.activityListeners {
			v(log)
		}
	}
	return err
}

func (admin *App) getHistory(resource *Resource, itemID int64) historyView {
	ret := historyView{}

	q := admin.Query()
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
		err := admin.Query().WhereIs("id", v.User).Get(&user)
		if err == nil {
			username = user.Name
			userurl = admin.GetURL(fmt.Sprintf("user/%d", user.ID))
		}

		activityURL := admin.GetURL(fmt.Sprintf("activitylog/%d", v.ID))
		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ID)

		ret.Items = append(ret.Items, historyItemView{
			ID:          v.ID,
			ActivityURL: activityURL,
			ActionType:  v.ActionType,
			ItemName:    itemName,
			ItemURL:     resource.GetURL(fmt.Sprintf("%d", v.ItemID)),
			UserName:    username,
			UserURL:     userurl,
			CreatedAt:   messages.Messages.Timestamp(user.Locale, v.CreatedAt, true),
		})
	}
	return ret
}

func initActivityLog(resource *Resource) {
	resource.CanView = permissionSysadmin
	resource.OrderDesc = true
	resource.HumanName = messages.Messages.GetNameFunction("admin_history")
}

func (admin App) createNewActivityLog(resource Resource, user User, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return admin.createActivityLog(ActivityLog{
		ResourceName: resource.ID,
		ItemID:       getItemID(item),
		ActionType:   "new",
		User:         user.ID,
		ContentAfter: string(data),
	})
}

func (admin App) createEditActivityLog(resource Resource, user User, itemID int64, before, after []byte) error {
	return admin.createActivityLog(ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "edit",
		User:          user.ID,
		ContentBefore: string(before),
		ContentAfter:  string(after),
	})
}

func (admin App) createDeleteActivityLog(resource Resource, user User, itemID int64, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return admin.createActivityLog(ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "delete",
		User:          user.ID,
		ContentBefore: string(data),
	})
}
