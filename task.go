package prago

import (
	"fmt"
	"mime/multipart"
	"runtime/debug"
	"time"
)

func (app *App) postInitTaskManager() {
	app.API("tasks/runtask").Method("POST").Permission(loggedPermission).Handler(func(request *Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")

		if request.csrfToken() != csrf {
			panic("wrong token")
		}

		task := app.tasksMap[id]
		if !request.Authorize(task.permission) {
			panic("not authorized")
		}
		app.runTask(task, request.UserID(), request.Locale(), request.Request().MultipartForm)
		request.Redirect(task.dashboard.board.getURL())
	})

	sysadminBoard.Dashboard(unlocalized("Cache")).AddTask(unlocalized("Delete cache"), "sysadmin", func(ta *TaskActivity) error {
		app.ClearCache()
		return nil
	})

}

type taskView struct {
	ID        string
	Name      string
	CSRFToken string
}

func (t *Task) taskView(locale, csrfToken string) taskView {
	return taskView{
		ID:        t.uuid,
		Name:      t.name(locale),
		CSRFToken: csrfToken,
	}
}

func (dashboard *Dashboard) getTasks(userID int64, userData UserData, csrfToken string) (ret []taskView) {
	for _, v := range dashboard.tasks {
		if userData.Authorize(v.permission) {
			ret = append(ret, v.taskView(userData.Locale(), csrfToken))
		}
	}
	return ret
}

// Task represent some user task
type Task struct {
	uuid        string
	name        func(string) string
	dashboard   *Dashboard
	permission  Permission
	handler     func(*TaskActivity) error
	lastStarted time.Time
}

// Task creates task
func (dashboard *Dashboard) AddTask(name func(string) string, permission Permission, handler func(*TaskActivity) error) {
	if dashboard.board.app.tasksMap == nil {
		dashboard.board.app.tasksMap = make(map[string]*Task)
	}

	if dashboard.board.app.validatePermission(permission) != nil {
		panic("invalid permission")
	}

	task := &Task{
		uuid:       randomString(20),
		name:       name,
		permission: sysadminPermission,
		dashboard:  dashboard,
		handler:    handler,
	}

	dashboard.tasks = append(dashboard.tasks, task)
	dashboard.board.app.tasksMap[task.uuid] = task

}

func (app *App) runTask(t *Task, userID int64, locale string, form *multipart.Form) {

	var name string = t.name(locale)

	var notification *Notification = app.Notification(name)
	notification.preName = t.dashboard.name(locale)

	activity := &TaskActivity{
		task:         t,
		notification: notification,
	}
	t.lastStarted = time.Now()

	notification.SetPrimaryAction("Ukončit", func() {
		activity.stoppedByUser = true
	})

	notification.disableCancel = true

	if userID > 0 {
		notification.Push(userID)
	}

	progress := -1.0
	notification.SetProgress(&progress)

	go func() {
		defer func() {
			notification.primaryAction = nil
			notification.secondaryAction = nil
			activity.notification.disableCancel = false
			activity.notification.progress = nil
			if r := recover(); r != nil {
				activity.notification.SetStyleFail()
				if activity.stoppedByUser {
					notification.SetDescription("Ukončeno uživatelem")
				} else {
					recoveryStr := fmt.Sprintf("%v, stack: %s", r, string(debug.Stack()))
					notification.SetDescription(recoveryStr)
				}
			} else {
				activity.notification.description = "Úspěšně dokončeno"
				activity.notification.SetStyleSuccess()
			}
		}()
		must(t.handler(activity))
	}()
}
