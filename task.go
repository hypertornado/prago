package prago

import (
	"fmt"
	"mime/multipart"
	"runtime/debug"
	"time"
)

type taskManager struct {
	app       *App
	tasksMap  map[string]*Task
	startedAt time.Time
}

func (app *App) preInitTaskManager() {
	app.taskManager = &taskManager{
		app:       app,
		tasksMap:  make(map[string]*Task),
		startedAt: time.Now(),
	}

}

func (app *App) postInitTaskManager() {
	go app.taskManager.startCRON()

	app.API("tasks/runtask").Method("POST").Permission(loggedPermission).Handler(func(request *Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")

		if request.csrfToken() != csrf {
			panic("wrong token")
		}

		task := app.taskManager.tasksMap[id]
		if !request.Authorize(task.permission) {
			panic("not authorize")
		}
		app.taskManager.run(task, request.UserID(), request.Locale(), request.Request().MultipartForm)

		fullURL := app.getAdminURL(task.dashboard.board.action.url)
		request.Redirect(fullURL)
	})

	sysadminBoard.Dashboard(unlocalized("Cache")).Task(unlocalized("Delete cache")).Handler(func(ta *TaskActivity) error {
		app.ClearCache()
		return nil
	})

}

func (tm *taskManager) startCRON() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, v := range tm.tasksMap {
				if v.cron > 0 {
					if tm.startedAt.Add(v.cron).Before(time.Now()) && v.lastStarted.Add(v.cron).Before(time.Now()) {
						tm.run(v, 0, "en", nil)
					}
				}
			}
		}
	}()
}

type taskView struct {
	ID        string
	Name      string
	CSRFToken string
}

func (t *Task) taskView(locale, csrfToken string) taskView {
	return taskView{
		ID:        t.id,
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
	id          string
	name        func(string) string
	dashboard   *Dashboard
	permission  Permission
	handler     func(*TaskActivity) error
	cron        time.Duration
	lastStarted time.Time
	//files       []*taskFileInput
}

// Task creates task
func (dashboard *Dashboard) Task(name func(string) string) *Task {

	id := randomString(20)
	_, ok := dashboard.board.app.taskManager.tasksMap[id]
	if ok {
		panic(fmt.Sprintf("Task '%s' already added.", id))
	}

	task := &Task{
		id:         id,
		name:       name,
		permission: sysadminPermission,
		dashboard:  dashboard,
	}

	dashboard.tasks = append(dashboard.tasks, task)
	dashboard.board.app.taskManager.tasksMap[task.id] = task

	return task
}

// Handler sets handler to task
func (t *Task) Handler(fn func(*TaskActivity) error) *Task {
	t.handler = fn
	return t
}

// SetPermission set permission to task
func (t *Task) Permission(permission string) *Task {
	t.permission = Permission(permission)
	return t
}

// RepeatEvery sets cron to task
func (t *Task) RepeatEvery(duration time.Duration) *Task {
	t.cron = duration
	return t
}

func (tm *taskManager) run(t *Task, userID int64, locale string, form *multipart.Form) {

	var name string = t.name(locale)

	var notification *Notification = tm.app.Notification(name)
	notification.preName = t.dashboard.name(locale)

	activity := &TaskActivity{
		task:         t,
		notification: notification,
		//files:        form,
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
