package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

//ALTER TABLE activitylog ADD INDEX idx_resource_item (resourcename, itemid);

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
	q := app.Query[activityLog]()
	if resource != nil {
		q.Is("ResourceName", resource.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}

	total, err := q.Count()
	Must(err)

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

	ret := app.NewTable()

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		ret.Row(Cell("Špatný formát stránkování"))
		return ret
	}

	q := app.Query[activityLog]()
	if resource != nil {
		q.Is("ResourceName", resource.getID())
	}
	if itemID > 0 {
		q.Is("ItemID", itemID)
	}

	q.In("user", request.Param("user"))

	total, err := q.Count()
	Must(err)

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

	ret.Header("Datum", "Uživatel", "Položka", "Úprava")

	for _, v := range items {
		var username, userurl string
		user := app.Query[user]().ID(v.User)
		locale := request.Locale()
		if user != nil {
			username = user.Name
			userurl = app.getAdminURL(fmt.Sprintf("user/%d", user.ID))
		}

		activityURL := app.getAdminURL(fmt.Sprintf("_activity?id=%d", v.ID))

		itemName := fmt.Sprintf("%s #%d", v.ResourceName, v.ItemID)
		var itemPreName string
		if resource != nil {
			item := resource.query(context.Background()).ID(v.ItemID)
			var name string
			if item != nil {
				name = resource.previewer(request, item).Name()
			}
			itemPreName = fmt.Sprintf("%s #%d", resource.singularName(request.Locale()), v.ItemID)
			itemName = name
		}

		actionCell := Cell(humanizeActionType(v.ActionType))

		actionCell.DescriptionBefore(fmt.Sprintf("#%d", v.ID))

		if v.ActionType == "delete" {
			actionCell.Red()
		}
		if v.ActionType == "new" {
			actionCell.Green()
		}
		if v.ActionType == "edit" {
			actionCell.Orange()

			diffLog := v.getInlineDiffLog(request)
			for _, item := range diffLog {
				actionCell.DescriptionAfter(item)
			}
		}

		actionCell.Button(&Button{
			Name:    "Zobrazit změnu",
			Icon:    iconView,
			OnClick: template.JS(fmt.Sprintf("popup(\"%s\")", activityURL)),
			//URL:     activityURL,
		})

		itemCell := Cell([2]string{resource.getURL(fmt.Sprintf("%d", v.ItemID)), itemName})
		itemCell.DescriptionBefore(itemPreName)

		dateCell := Cell(messages.Timestamp(locale, v.CreatedAt, true))
		userCell := Cell([2]string{userurl, username})

		ret.Row(
			dateCell,
			userCell,
			itemCell,
			actionCell,
		)

	}
	return ret
}

func humanizeActionType(in string) string {
	if in == "delete" {
		return "Smazat"
	}
	if in == "new" {
		return "Vytvořit"
	}
	if in == "edit" {
		return "Upravit"
	}
	return in
}

func humanizeActionTypePast(in string) string {
	if in == "delete" {
		return "Smazal"
	}
	if in == "new" {
		return "Vytvořil"
	}
	if in == "edit" {
		return "Upravil"
	}
	return in
}

func (app *App) initActivityLog() {

	app.PopupForm("_activity", func(form *Form, request *Request) {
		id := request.Param("id")
		form.AddHidden("id").SetValue(id)
		form.AutosubmitFirstTime = true
	}, func(fv FormValidation, request *Request) {
		table := app.NewTable()

		activity := app.Query[activityLog]().ID(request.Param("id"))
		if activity == nil {
			fv.AfterContent("Activity not found")
			return
		}

		resource := app.getResourceByID(activity.ResourceName)
		if !request.Authorize(resource.canView) {
			fv.AfterContent("Not allowed")
			return
		}

		item := resource.query(context.Background()).ID(activity.ItemID)
		var itemName string
		if item != nil {
			itemName = resource.previewer(request, item).Name()
		} else {
			itemName = fmt.Sprintf("#%d", activity.ItemID)
		}

		//table.Row(Cell("Editace").Header(), Cell(fmt.Sprintf("#%d", activity.ID)).Colspan(2).URL(fmt.Sprintf("/admin/activitylog/%d", activity.ID)))
		//table.Row(Cell("Tabulka").Header(), Cell(resource.pluralName(request.Locale())).Colspan(2).URL(resource.getURL("")))
		table.Row(Cell(resource.singularName(request.Locale())).Header(), Cell(itemName).Colspan(2).URL(resource.getURL(fmt.Sprintf("%d", activity.ItemID))))
		//table.Row(Cell("Typ akce").Header(), Cell(activity.ActionType).Colspan(2))
		//table.Row(Cell("Upraveno").Header(), Cell(activity.CreatedAt.Format("2. 1. 2006 15:04:05")).Colspan(2))
		//Cell(app.GetUserName(activity.User)).Colspan(2).URL(fmt.Sprintf("/admin/user/%d", activity.User))
		userCell := Cell(app.GetUserName(activity.User)).Colspan(2).URL(fmt.Sprintf("/admin/user/%d", activity.User))
		userCell.DescriptionAfter(activity.CreatedAt.Format("2. 1. 2006 15:04:05"))
		userCell.DescriptionAfter(fmt.Sprintf("Úprava #%d", activity.ID))

		table.Row(Cell(humanizeActionTypePast(activity.ActionType)).Header(), userCell)

		fromMap := getDiffMap(activity.ContentBefore)
		toMap := getDiffMap(activity.ContentAfter)

		table.AddTable()

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

	}).Permission("logged").Name(Unlocalized("Úprava")).Icon(iconActivity)

	app.activityLogResource = app.NewResource[activityLog]()
	app.activityLogResource.Board(sysadminBoard)
	app.activityLogResource.icon = iconActivity
	app.activityLogResource.canView = Permission(sysadminRoleName)
	app.activityLogResource.canUpdate = Permission("nobody")
	app.activityLogResource.canDelete = Permission("nobody")
	app.activityLogResource.orderDesc = true
	app.activityLogResource.Name(messages.GetNameFunction("history"), messages.GetNameFunction("history"))

	app.PreviewURLFunction(func(activityLog *activityLog) string {
		return fmt.Sprintf("/admin/_activity?id=%d", activityLog.ID)
	})

}

func (activity *activityLog) getInlineDiffLog(request *Request) (ret []string) {
	app := request.app
	resource := app.getResourceByID(activity.ResourceName)
	if resource == nil {
		return nil
	}

	fromMap := getDiffMap(activity.ContentBefore)
	toMap := getDiffMap(activity.ContentAfter)

	for _, field := range resource.fields {
		if !request.Authorize(field.canView) {
			continue
		}

		fromContent := fromMap[field.id]
		toContent := toMap[field.id]

		if fromContent != toContent {
			ret = append(ret, fmt.Sprintf("%s: %s → %s", field.name(request.Locale()), fromContent, toContent))
		}

	}
	return
}

func getDiffMap(inStr string) map[string]string {
	var objectMap map[string]any
	var ret = map[string]string{}
	if inStr == "" {
		return ret
	}
	Must(json.Unmarshal([]byte(inStr), &objectMap))

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

	if resource.activityLog {
		err = resource.app.Create(log)
	}

	if err == nil {
		for _, v := range resource.app.activityListeners {
			v(log.activity())
		}
	}
	return err

}
