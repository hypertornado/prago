package prago

import (
	"encoding/json"
	"errors"
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

// ListenActivity listens to all changes in app's administration
func (app *App) ListenActivity(handler func(Activity)) {
	app.activityListeners = append(app.activityListeners, handler)
}

func (app *App) getHistory(resourceData *resourceData, itemID int64) historyView {
	ret := historyView{}

	q := app.activityLogResource.Query()
	if resourceData != nil {
		q.Is("ResourceName", resourceData.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}
	q.Limit(250)
	q.OrderDesc("ID")

	items := q.List()

	for _, v := range items {
		var username, userurl string
		user := app.UsersResource.ID(v.User)
		locale := "en"
		if user != nil {
			username = user.Name
			userurl = app.getAdminURL(fmt.Sprintf("user/%d", user.ID))
			locale = user.Locale
		}

		activityURL := app.getAdminURL(fmt.Sprintf("activitylog/%d", v.ID))
		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ID)

		ret.Items = append(ret.Items, historyItemView{
			ID:          v.ID,
			ActivityURL: activityURL,
			ActionType:  v.ActionType,
			ItemName:    itemName,
			ItemURL:     resourceData.getURL(fmt.Sprintf("%d", v.ItemID)),
			UserName:    username,
			UserURL:     userurl,
			CreatedAt:   messages.Timestamp(locale, v.CreatedAt, true),
		})
	}
	return ret
}

func (app *App) initActivityLog() {
	app.activityLogResource = NewResource[activityLog](app)
	app.activityLogResource.data.canView = Permission(sysadminRoleName)
	app.activityLogResource.data.orderDesc = true
	app.activityLogResource.Name(messages.GetNameFunction("admin_history"), messages.GetNameFunction("admin_history"))
}

func (resourceData *resourceData) LogActivity(user *user, before, after any) error {
	var activityType string
	switch {
	case before == nil && after != nil:
		activityType = "new"
	case before != nil && after != nil:
		activityType = "edit"
	case before != nil && after == nil:
		activityType = "delete"
	default:
		return errors.New("unknown activity type")

	}

	var itemID int64 = -1
	if before != nil {
		itemID = getItemID(before)
	} else {
		itemID = getItemID(after)
	}

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

	log := &activityLog{
		ResourceName:  resourceData.id,
		ItemID:        itemID,
		ActionType:    activityType,
		User:          user.ID,
		ContentBefore: string(beforeData),
		ContentAfter:  string(afterData),
	}

	err = resourceData.app.activityLogResource.Create(log)
	if err == nil {
		for _, v := range resourceData.app.activityListeners {
			v(log.activity())
		}
	}
	return err

}
