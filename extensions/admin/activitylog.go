package admin

import (
	"encoding/json"
	"fmt"
	"time"
)

type HistoryView struct {
	Items []HistoryItemView
}

type HistoryItemView struct {
	ID          int64
	ActionType  string
	ActivityURL string
	ItemName    string
	ItemURL     string
	UserName    string
	UserURL     string
	CreatedAt   string
}

func (admin *Admin) getHistory(resource *Resource, user int64, itemID int64) HistoryView {
	ret := HistoryView{}

	q := admin.Query()
	if resource != nil {
		q.WhereIs("ResourceName", resource.ID)
	}
	if user < 0 {
		q.WhereIs("User", user)
	}
	if itemID > 0 {
		q.WhereIs("ItemID", itemID)
	}
	q.Limit(250)
	q.OrderDesc("ID")

	var items []*ActivityLog
	err := q.Get(&items)
	if err != nil {
		panic(err)
	}

	for _, v := range items {
		var username, userurl string

		var user User
		err := admin.Query().WhereIs("id", v.User).Get(&user)
		if err == nil {
			username = user.Name
			userurl = fmt.Sprintf("%s/user/%d", admin.Prefix, user.ID)
		}

		activityURL := fmt.Sprintf("%s/activitylog/%d", admin.Prefix, v.ID)

		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ID)

		ret.Items = append(ret.Items, HistoryItemView{
			ID:          v.ID,
			ActivityURL: activityURL,
			ActionType:  v.ActionType,
			ItemName:    itemName,
			ItemURL:     admin.getURL(resource, fmt.Sprintf("%d", v.ItemID)),
			UserName:    username,
			UserURL:     userurl,
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return ret
}

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

func initActivityLog(resource *Resource) {
	resource.OrderDesc = true
}

func (ActivityLog) Authenticate(u *User) bool {
	return AuthenticateSysadmin(u)
}

func (admin Admin) createNewActivityLog(resource Resource, user User, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	log := ActivityLog{
		ResourceName: resource.ID,
		ItemID:       getItemID(item),
		ActionType:   "new",
		User:         user.ID,
		ContentAfter: string(data),
	}
	return admin.Create(&log)
}

func (admin Admin) createEditActivityLog(resource Resource, user User, itemID int64, before, after []byte) error {
	log := ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "edit",
		User:          user.ID,
		ContentBefore: string(before),
		ContentAfter:  string(after),
	}
	return admin.Create(&log)
}

func (admin Admin) createDeleteActivityLog(resource Resource, user User, itemID int64, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	log := ActivityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "delete",
		User:          user.ID,
		ContentBefore: string(data),
	}
	return admin.Create(&log)
}
