package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

type activityLog struct {
	ID            int64
	ResourceName  string
	ItemID        int64
	ActionType    string
	User          int64  `prago-type:"relation"`
	ContentBefore string `prago-type:"text"`
	ContentAfter  string `prago-type:"text"`
	CreatedAt     time.Time
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

// ListenActivity listens to all changes in app's administration
func (app *App) ListenActivity(handler func(Activity)) {
	app.activityListeners = append(app.activityListeners, handler)
}

var activityItemsPerPage int64 = 100

func (app *App) getHistorySelect(request *Request, resource *Resource, itemID int64) (ret [][2]string) {
	q := Query[activityLog](app)
	if resource != nil {
		q.Is("ResourceName", resource.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}

	total, err := q.Count()
	must(err)

	pages := total / activityItemsPerPage
	for i := 0; i <= int(pages); i++ {
		ret = append(ret, [2]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("Stránka %d", i+1),
		})
	}
	return ret
}

func (app *App) getHistoryTable(request *Request, resource *Resource, itemID int64, pageStr string) *Table {

	ret := app.Table()

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		ret.Row(Cell("Špatný formát stránkování"))
		return ret
	}

	q := Query[activityLog](app)
	if resource != nil {
		q.Is("ResourceName", resource.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}

	total, err := q.Count()
	must(err)

	offset := activityItemsPerPage * (int64(page) - 1)

	q.Offset(offset)
	q.Limit(activityItemsPerPage)
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

		activityURL := app.getAdminURL(fmt.Sprintf("_activity?id=%d", v.ID))

		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ItemID)
		if resource != nil {
			item := resource.query(context.Background()).ID(v.ItemID)
			var name string
			if item != nil {
				name = resource.previewer(request, item).Name()
			}
			itemName = fmt.Sprintf("#%d %s", v.ItemID, name)
		}

		ret.Row(
			Cell([2]string{activityURL, fmt.Sprintf("%d", v.ID)}),
			Cell(v.ActionType),
			Cell([2]string{resource.getURL(fmt.Sprintf("%d", v.ItemID)), itemName}),
			Cell([2]string{userurl, username}),
			Cell(messages.Timestamp(locale, v.CreatedAt, true)),
		)

	}
	return ret
}

func (app *App) initActivityLog() {

	ActionForm(app, "_activity", func(form *Form, request *Request) {
		id := request.Param("id")
		form.AddTextInput("id", "ID úpravy").Value = id
		form.AddSubmit("Zobrazit")

		if id != "" {
			form.AutosubmitFirstTime = true
		}
	}, func(fv FormValidation, request *Request) {
		table := app.Table()

		activity := Query[activityLog](app).ID(request.Param("id"))
		if activity == nil {
			fv.AfterContent("Activity not found")
			return
		}

		resource := app.getResourceByID(activity.ResourceName)
		if !request.Authorize(resource.canView) {
			fv.AfterContent("Not allowed")
			return
		}

		table.Row(Cell("Editace:").Header(), Cell(fmt.Sprintf("#%d", activity.ID)).Colspan(2).URL(fmt.Sprintf("/admin/activitylog/%d", activity.ID)))
		table.Row(Cell("Tabulka:").Header(), Cell(resource.pluralName(request.Locale())).Colspan(2).URL(resource.getURL("")))
		table.Row(Cell("Položka:").Header(), Cell(fmt.Sprintf("#%d", activity.ItemID)).Colspan(2).URL(resource.getURL(fmt.Sprintf("%d", activity.ItemID))))
		table.Row(Cell("Typ akce:").Header(), Cell(activity.ActionType).Colspan(2))
		table.Row(Cell("Upraveno:").Header(), Cell(activity.CreatedAt.Format("2006-01-02 15:04:05")).Colspan(2))

		user := Query[user](app).ID(activity.User)
		if user != nil {
			var userNames []string
			userNames = append(userNames, fmt.Sprintf("#%d", user.ID))
			if user.Username != "" {
				userNames = append(userNames, user.Username)
			}
			if user.Name != "" {
				userNames = append(userNames, user.Name)
			}
			if user.Email != "" {
				userNames = append(userNames, user.Email)
			}
			table.Row(Cell("Upraveno uživatelem:").Header(), Cell(strings.Join(userNames, " · ")).Colspan(2).URL(fmt.Sprintf("/admin/user/%d", user.ID)))
		}

		fromMap := getDiffMap(activity.ContentBefore)
		toMap := getDiffMap(activity.ContentAfter)

		table.Header("Jméno pole", "Před", "Po")

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}

			fromContent := fromMap[field.id]
			toContent := toMap[field.id]

			nameCell := Cell(field.name(request.Locale()))
			cellFrom := Cell(fromContent)
			cellTo := Cell(toContent)

			isSame := fromContent == toContent
			if isSame {
				cellFrom.Green()
				cellTo.Green()
			} else {
				cellFrom.Orange()
				cellTo.Orange()
			}

			table.Row(nameCell, cellFrom, cellTo)
		}

		fv.AfterContent(table.ExecuteHTML())

	}).Permission("logged").Name(unlocalized("Úpravy")).Icon(iconActivity)

	app.activityLogResource = NewResource[activityLog](app)
	app.activityLogResource.Board(sysadminBoard)
	app.activityLogResource.icon = iconActivity
	app.activityLogResource.canView = Permission(sysadminRoleName)
	app.activityLogResource.canUpdate = Permission("nobody")
	app.activityLogResource.canDelete = Permission("nobody")
	app.activityLogResource.orderDesc = true
	app.activityLogResource.Name(messages.GetNameFunction("admin_history"), messages.GetNameFunction("admin_history"))

	PreviewURLFunction(app, func(activityLog *activityLog) string {
		return fmt.Sprintf("/admin/_activity?id=%d", activityLog.ID)
	})

}

func getDiffMap(inStr string) map[string]string {
	var objectMap map[string]interface{}
	var ret = map[string]string{}
	if inStr == "" {
		return ret
	}
	must(json.Unmarshal([]byte(inStr), &objectMap))

	for k, v := range objectMap {
		ret[strings.ToLower(k)] = fmt.Sprintf("%v", v)
	}
	return ret
}

func (resource *Resource) logActivity(userData UserData, before, after any) error {
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
		itemID = resource.previewer(userData, before).ID()
	} else {
		itemID = resource.previewer(userData, after).ID()
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
		ResourceName:  resource.id,
		ItemID:        itemID,
		ActionType:    activityType,
		User:          userData.UserID(),
		ContentBefore: string(beforeData),
		ContentAfter:  string(afterData),
	}

	err = CreateItem(resource.app, log)
	if err == nil {
		for _, v := range resource.app.activityListeners {
			v(log.activity())
		}
	}
	return err

}
