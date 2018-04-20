package administration

import (
	"encoding/json"
	"fmt"
	"time"
)

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

func (admin *Administration) getHistory(resource *Resource, user int64, itemID int64) historyView {
	ret := historyView{}

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

	var items []*activityLog
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
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return ret
}

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

func initActivityLog(resource *Resource) {
	resource.CanView = permissionSysadmin
	resource.OrderDesc = true
}

func (admin Administration) createNewActivityLog(resource Resource, user User, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	log := activityLog{
		ResourceName: resource.ID,
		ItemID:       getItemID(item),
		ActionType:   "new",
		User:         user.ID,
		ContentAfter: string(data),
	}
	return admin.Create(&log)
}

func (admin Administration) createEditActivityLog(resource Resource, user User, itemID int64, before, after []byte) error {
	log := activityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "edit",
		User:          user.ID,
		ContentBefore: string(before),
		ContentAfter:  string(after),
	}
	return admin.Create(&log)
}

func (admin Administration) createDeleteActivityLog(resource Resource, user User, itemID int64, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	log := activityLog{
		ResourceName:  resource.ID,
		ItemID:        itemID,
		ActionType:    "delete",
		User:          user.ID,
		ContentBefore: string(data),
	}
	return admin.Create(&log)
}

func (admin Administration) createExportActivityLog(resource Resource, user User, item exportFormData) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	log := activityLog{
		ResourceName:  resource.ID,
		ActionType:    "export",
		User:          user.ID,
		ContentBefore: string(data),
	}
	return admin.Create(&log)
}
