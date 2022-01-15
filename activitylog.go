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

type Activity struct {
	ID           int64
	ResourceID   string
	User         int64
	ActivityType string
}

func (al activityLog) activity() Activity {
	return Activity{
		ID:           al.ItemID,
		ResourceID:   al.ResourceName,
		User:         al.User,
		ActivityType: al.ActionType,
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
func (app *App) ListenActivity(handler func(Activity)) {
	app.activityListeners = append(app.activityListeners, handler)
}

func (app *App) getHistory(resource *resource, itemID int64) historyView {
	ret := historyView{}

	q := GetResource[activityLog](app).Query()
	if resource != nil {
		q.Is("ResourceName", resource.id)
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}
	q.Limit(250)
	q.OrderDesc("ID")

	items := q.List()

	for _, v := range items {
		var username, userurl string
		user := app.UsersResource.Is("id", v.User).First()
		if user != nil {
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

func initActivityLog(resource *resource) {
	resource.canView = Permission(sysadminRoleName)
	resource.orderDesc = true
	resource.name = messages.GetNameFunction("admin_history")
}

func (app App) LogActivity(activityType string, userID int64, resourceID string, itemID int64, before, after interface{}) error {
	var err error

	var beforeData []byte
	if before != nil {
		beforeData, err = json.Marshal(before)
		if err != nil {
			return fmt.Errorf("can't marshal before data: %s", err)
		}
	}

	var afterData []byte
	if after != nil {
		afterData, err = json.Marshal(after)
		if err != nil {
			return fmt.Errorf("can't marshal after data: %s", err)
		}
	}

	log := activityLog{
		ResourceName:  resourceID,
		ItemID:        itemID,
		ActionType:    activityType,
		User:          userID,
		ContentBefore: string(beforeData),
		ContentAfter:  string(afterData),
	}

	err = app.create(&log)
	if err == nil {
		for _, v := range app.activityListeners {
			v(log.activity())
		}
	}
	return err

}
