package prago

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"sync"
	"time"
)

type conflictResponse struct {
	Show bool
	Text string
}

type conflictQueueItem struct {
	ResourceID string
	ItemID     int64
	UserID     int64
	CreatedAt  time.Time
}

func (app *App) initResourceConflict() {

	app.conflictMutex = &sync.Mutex{}

	app.API("_conflict").HandlerJSON(func(request *Request) any {
		//return &conflictResponse{}
		version, err := strconv.Atoi(request.Param("version"))
		if err != nil {
			return &conflictResponse{
				Show: true,
				Text: fmt.Sprintf("Nesprávný formát verze: '%s'", request.Param("version")),
			}
		}
		return app.getConflictResponse(request, int64(version))
	}).Method("POST").Permission(loggedPermission)

	ActionUI(app, "_conflicts", func(request *Request) template.HTML {

		conflicts := app.getCurrentConflicts()

		table := app.Table()

		table.Header("Resource", "Item", "User", "Date")
		for _, conflict := range conflicts {
			table.Row(
				Cell(conflict.ResourceID).URL(fmt.Sprintf("/admin/%s", conflict.ResourceID)),
				Cell(conflict.ItemID).URL(fmt.Sprintf("/admin/%s/%d", conflict.ResourceID, conflict.ItemID)),
				Cell(conflict.UserID).URL(fmt.Sprintf("/admin/user/%d", conflict.UserID)),
				Cell(conflict.CreatedAt.Format("15:04:05")),
			)
		}
		return table.ExecuteHTML()

	}).Permission(sysadminPermission).Name(unlocalized("Conflict checks")).Board(sysadminBoard)

}

func (app *App) getConflictResponse(request *Request, version int64) *conflictResponse {
	if version <= 0 {
		return &conflictResponse{}
	}

	logItem := Query[activityLog](app).ID(version)
	if logItem == nil {
		return &conflictResponse{
			Show: true,
			Text: fmt.Sprintf("Tuto verzi nelze najít: '%d'", version),
		}
	}
	currentVersion := Query[activityLog](app).Is("resourcename", logItem.ResourceName).Is("ItemID", logItem.ItemID).OrderDesc("id").First()
	if currentVersion.ID != logItem.ID {
		if request.Authorize(app.UsersResource.canView) {
			//logItem := Query[activityLog](app).ID(currentVersion)

			var userName = fmt.Sprintf("#%d", currentVersion.User)
			ud := app.GetUserData(currentVersion.User)
			if ud != nil {
				userName = ud.Name()
			}
			return &conflictResponse{
				Show: true,
				Text: fmt.Sprintf("Editační konflikt. Tuto položku upravil uživatel '%s' v čase %s. Obnovte prosím stránku", userName, logItem.CreatedAt.Format("15:04:05")),
			}
		} else {
			return &conflictResponse{
				Show: true,
				Text: "Editační konflikt. Někdo jiný upravil tuto položku před vámi. Obnovte prosím stránku",
			}
		}
	}
	return app.markConflict(request, logItem.ResourceName, logItem.ItemID)
}

func (app *App) getCurrentConflicts() []*conflictQueueItem {
	app.conflictMutex.Lock()
	defer app.conflictMutex.Unlock()
	return app.conflictQueue
}

func (app *App) markConflict(request *Request, resourceID string, itemID int64) *conflictResponse {
	app.conflictMutex.Lock()
	defer app.conflictMutex.Unlock()

	var copy []*conflictQueueItem

	timeLimit := time.Now().Add(-10 * time.Second)

	currentUser := request.UserID()

	usersMap := map[int64]bool{}

	for _, item := range app.conflictQueue {
		if item.CreatedAt.Before(timeLimit) {
			continue
		}
		copy = append(copy, item)

		if item.ResourceID == resourceID && item.ItemID == itemID && item.UserID != currentUser {
			usersMap[item.UserID] = true
		}
	}

	newItem := &conflictQueueItem{
		ResourceID: resourceID,
		ItemID:     itemID,
		UserID:     request.UserID(),
		CreatedAt:  time.Now(),
	}
	copy = append(copy, newItem)
	app.conflictQueue = copy

	if len(usersMap) == 0 {
		return &conflictResponse{}
	}

	if !request.Authorize(app.UsersResource.canView) {
		return &conflictResponse{
			Show: true,
			Text: "Tuto položku editují i jiní uživatelé",
		}
	}

	var userNames []string
	for userID := range usersMap {
		var userName = fmt.Sprintf("#%d", userID)
		ud := app.GetUserData(userID)
		if ud != nil {
			userName = ud.Name()
		}
		userNames = append(userNames, userName)
	}

	if len(userNames) == 1 {
		return &conflictResponse{
			Show: true,
			Text: "Tuto položku edituje také uživatel " + userNames[0],
		}
	}

	return &conflictResponse{
		Show: true,
		Text: "Tuto položku editují uživatelé: " + strings.Join(userNames, ", "),
	}
}

func GetItemVersion[T any](request *Request, item *T) int64 {
	app := request.app
	resource := getResource[T](app)
	id := resource.previewer(request, item).ID()
	return resource.currentItemVersion(id)
}

func (resource *Resource) currentItemVersion(itemID int64) int64 {
	app := resource.app
	lastLog := Query[activityLog](app).Is("resourcename", resource.id).Is("ItemID", itemID).OrderDesc("id").First()
	if lastLog != nil {
		return lastLog.ID
	}
	return 0
}

func (resource *Resource) validateConflict(request *Request, vc *itemValidation, itemID int64) {
	app := resource.app
	itemVersionStr := request.Param("_itemversion")
	if itemVersionStr == "" || itemVersionStr == "0" {
		return
	}
	oldVersion, err := strconv.Atoi(itemVersionStr)
	must(err)

	currentVersion := resource.currentItemVersion(itemID)

	if currentVersion != int64(oldVersion) {

		if request.Authorize(app.UsersResource.canView) {
			logItem := Query[activityLog](app).ID(currentVersion)

			var userName = fmt.Sprintf("#%d", logItem.User)
			ud := app.GetUserData(logItem.User)
			if ud != nil {
				userName = ud.Name()
			}
			vc.AddError(fmt.Sprintf("Editační konflikt. Tuto položku upravil uživatel '%s' v čase %s. Obnovte prosím stránku", userName, logItem.CreatedAt.Format("15:04:05")))

		} else {
			vc.AddError("Editační konflikt. Někdo jiný upravil tuto položku před vámi. Obnovte prosím stránku")
		}
	}

}
