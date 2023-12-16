package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"
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

func (app *App) getHistoryTable(request *Request, resourceData *Resource, itemID int64, pageStr string) *Table {

	ret := app.Table()

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		ret.Row(Cell("Špatný formát stránkování"))
		return ret
	}

	q := Query[activityLog](app)
	//q := app.activityLogResource.Query(context.Background())
	if resourceData != nil {
		q.Is("ResourceName", resourceData.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}

	var itemsPerPage int64 = 100

	total, err := q.Count()
	must(err)

	offset := itemsPerPage * (int64(page) - 1)

	q.Offset(offset)
	q.Limit(itemsPerPage)
	q.OrderDesc("ID")

	items := q.List()

	if len(items) > 0 {
		ret.AddFooterText(fmt.Sprintf("Zobrazuji položku %d - %d z celkem %d položek.", offset+1, offset+int64(len(items)), total))
	} else {
		if total > 0 {
			ret.AddFooterText(fmt.Sprintf("Na stránce %d nic není. Celkem %d položek.", page, total))
		} else {
			ret.AddFooterText("Žádná úprava nenalezena.")
		}
		return ret

	}

	ret.Header("#", "Typ akce", "Položka", "Uživatel", "Datum")

	for _, v := range items {
		var username, userurl string
		user := Query[user](app).ID(v.User)
		locale := "en"
		if user != nil {
			username = user.Name
			userurl = app.getAdminURL(fmt.Sprintf("user/%d", user.ID))
			locale = user.Locale
		}

		activityURL := app.getAdminURL(fmt.Sprintf("activitylog/%d", v.ID))

		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ItemID)
		if resourceData != nil {
			item := resourceData.query(context.Background()).ID(v.ItemID)
			var name string
			if item != nil {
				name = resourceData.previewer(request, item).Name()
			}
			itemName = fmt.Sprintf("#%d %s", v.ItemID, name)
		}

		ret.Row(
			Cell([2]string{activityURL, fmt.Sprintf("%d", v.ID)}),
			Cell(v.ActionType),
			Cell([2]string{resourceData.getURL(fmt.Sprintf("%d", v.ItemID)), itemName}),
			Cell([2]string{userurl, username}),
			Cell(messages.Timestamp(locale, v.CreatedAt, true)),
		)

	}
	return ret
}

func (app *App) initActivityLog() {
	app.activityLogResource = NewResource[activityLog](app)
	app.activityLogResource.Board(sysadminBoard)
	app.activityLogResource.icon = "glyphicons-basic-58-history.svg"
	app.activityLogResource.canView = Permission(sysadminRoleName)
	app.activityLogResource.orderDesc = true
	app.activityLogResource.Name(messages.GetNameFunction("admin_history"), messages.GetNameFunction("admin_history"))
}

func (resourceData *Resource) logActivity(request *Request, before, after any) error {
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
		itemID = resourceData.previewer(request, before).ID()
	} else {
		itemID = resourceData.previewer(request, after).ID()
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
		User:          request.UserID(),
		ContentBefore: string(beforeData),
		ContentAfter:  string(afterData),
	}

	err = CreateItem(resourceData.app, log)
	if err == nil {
		for _, v := range resourceData.app.activityListeners {
			v(log.activity())
		}
	}
	return err

}
